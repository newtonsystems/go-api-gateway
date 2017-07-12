package main

import (
	//"context"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	//"golang.org/x/net/context"
	//"github.com/apache/thrift/lib/go/thrift"
	"github.com/go-kit/kit/log/term"
	lightstep "github.com/lightstep/lightstep-tracer-go"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"

	"github.com/go-kit/kit/endpoint"
	"github.com/newtonsystems/go-api-gateway/app"

	//"go-hello/app/pb"
	//"github.com/go-kit/kit/examples/addsvc/pb"
	//thriftadd "go-hello/app/cmd/addsvc/thrift/gen-go/addsvc"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/tracing/opentracing"

	"github.com/newtonsystems/grpc_types/go/grpc_types"
)

var ColourKeys = func(keyvals ...interface{}) term.FgBgColor {
	for _, item := range keyvals {
		if item == "msg.ERROR" {
			return term.FgBgColor{Fg: term.DarkRed, Bg: term.White}
		} else if item == "msg.WARNING" {
			return term.FgBgColor{Fg: term.Yellow, Bg: term.White}
		}
	}

	return term.FgBgColor{}
}

func main() {

	var (
		debugAddr = flag.String("debug.addr", ":8080", "Debug and metrics listen address")
		//httpAddr  = flag.String("http.addr", ":8081", "HTTP listen address")
		//grpcAddr  = flag.String("grpc.addr", ":8042", "gRPC (HTTP) listen address")

		debugAnyGRPCService = flag.Bool("debug.grpc.any", true, "true to enable access to any grpc service (NEVER SET TO TRUE USE IN PRODUCTION)")
		//thriftAddr       = flag.String("thrift.addr", ":8083", "Thrift listen address")
		//thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		//thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		//thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
		zipkinAddr      = flag.String("zipkin.addr", "", "Enable Zipkin tracing via a Zipkin HTTP Collector endpoint")
		zipkinKafkaAddr = flag.String("zipkin.kafka.addr", "", "Enable Zipkin tracing via a Kafka server host:port")
		appdashAddr     = flag.String("appdash.addr", "", "Enable Appdash tracing via an Appdash server host:port")
		lightstepToken  = flag.String("lightstep.token", "", "Enable LightStep tracing via a LightStep access token")

		// Connect to linkerd ingress
		//linkerdAddr = flag.String("linkerd.addr", ":4041", "Linkerd ingress address")

		// Debug only (Should NEVER be used in production)
		httpAnyServiceAddr = flag.String("debug.httpanyservice.addr", ":9001", "HTTP listen address for accessing any service")
		gRPCAnyServiceAddr = flag.String("debug.grpcanyservice.addr", ":9002", "gRPC (HTTP) listen address for accessing any service")
	)
	flag.Parse()

	// Color by level value
	colorFn := func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] != "level" {
				continue
			}
			switch keyvals[i+1] {
			case "debug":
				return term.FgBgColor{Fg: term.DarkGray}
			case "info":
				return term.FgBgColor{Fg: term.DarkGreen}
			case "warn":
				return term.FgBgColor{Fg: term.Yellow, Bg: term.White}
			case "error":
				return term.FgBgColor{Fg: term.Red}
			case "crit":
				return term.FgBgColor{Fg: term.Gray, Bg: term.DarkRed}
			default:
				return term.FgBgColor{}
			}
		}
		return term.FgBgColor{}
	}

	// Logging domain.
	var logger log.Logger
	{
		//logger = log.NewLogfmtLogger(os.Stdout)
		logger = term.NewLogger(os.Stdout, log.NewLogfmtLogger, colorFn)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	logger.Log("msg", "hello", "level", "info")
	defer logger.Log("msg", "goodbye")

	stdlog.SetOutput(log.NewStdlibAdapter(logger))
	stdlog.Print("I sure like pie")

	// Metrics domain.
	// var ints, chars metrics.Counter
	// {
	// 	// Business level metrics.
	// 	ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 		Namespace: "addsvc",
	// 		Name:      "integers_summed",
	// 		Help:      "Total count of integers summed via the Sum method.",
	// 	}, []string{})
	// 	chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
	// 		Namespace: "addsvc",
	// 		Name:      "characters_concatenated",
	// 		Help:      "Total count of characters concatenated via the Concat method.",
	// 	}, []string{})
	// }
	var duration metrics.Histogram
	{
		// Transport level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "addsvc",
			Name:      "request_duration_ns",
			Help:      "Request duration in nanoseconds.",
		}, []string{"method", "success"})
	}

	// Tracing domain.
	var tracer stdopentracing.Tracer
	{
		if *zipkinAddr != "" {
			logger := log.With(logger, "tracer", "ZipkinHTTP")
			logger.Log("addr", *zipkinAddr)

			// endpoint typically looks like: http://zipkinhost:9411/api/v1/spans
			collector, err := zipkin.NewHTTPCollector(*zipkinAddr)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			defer collector.Close()

			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "localhost:80", "addsvc"),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else if *zipkinKafkaAddr != "" {
			logger := log.With(logger, "tracer", "ZipkinKafka")
			logger.Log("addr", *zipkinKafkaAddr)

			collector, err := zipkin.NewKafkaCollector(
				strings.Split(*zipkinKafkaAddr, ","),
				zipkin.KafkaLogger(log.NewNopLogger()),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			defer collector.Close()

			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "localhost:80", "addsvc"),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else if *appdashAddr != "" {
			logger := log.With(logger, "tracer", "Appdash")
			logger.Log("addr", *appdashAddr)
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
		} else if *lightstepToken != "" {
			logger := log.With(logger, "tracer", "LightStep")
			logger.Log() // probably don't want to print out the token :)
			tracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: *lightstepToken,
			})
			defer lightstep.FlushLightStepTracer(tracer)
		} else {
			logger := log.With(logger, "tracer", "none")
			logger.Log()
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	// Endpoint domain.

	// Mechanical domain.
	errc := make(chan error)
	//ctx := context.Background()

	conn, err := grpc.Dial("192.168.99.100:30000", grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		logger.Log("msg", "Failed to connect to linkerd", "level", "crit")
		errc <- err
		return
	}
	defer conn.Close()

	var sayHelloEndpoint endpoint.Endpoint
	{
		sayHelloDuration := duration.With("method", "SayHello")
		sayHelloLogger := log.With(logger, "method", "SayHello")

		sayHelloEndpoint = addsvc.MakeSayHelloEndpoint(conn)
		sayHelloEndpoint = opentracing.TraceServer(tracer, "SayHello")(sayHelloEndpoint)
		sayHelloEndpoint = addsvc.EndpointInstrumentingMiddleware(sayHelloDuration)(sayHelloEndpoint)
		sayHelloEndpoint = addsvc.EndpointLoggingMiddleware(sayHelloLogger)(sayHelloEndpoint)
	}

	endpoints := addsvc.Endpoints{
		SayHelloEndpoint: sayHelloEndpoint,
	}

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Debug listener.
	go func() {
		logger := log.With(logger, "transport", "debug")

		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		m.Handle("/metrics", promhttp.Handler())

		logger.Log("addr", *debugAddr)
		errc <- http.ListenAndServe(*debugAddr, m)
	}()

	// HTTP transport.
	// go func() {
	// 	logger := log.With(logger, "transport", "HTTP")
	// 	h := addsvc.MakeHTTPHandler(endpoints, tracer, logger)
	// 	logger.Log("addr", *httpAddr)
	// 	errc <- http.ListenAndServe(*httpAddr, h)
	// }()

	// gRPC transport.
	// go func() {
	// 	logger := log.With(logger, "transport", "gRPC")

	// 	ln, err := net.Listen("tcp", *grpcAddr)
	// 	if err != nil {
	// 		errc <- err
	// 		return
	// 	}

	// 	srv := addsvc.MakeGRPCServer(endpoints, tracer, logger)
	// 	s := grpc.NewServer()
	// 	pb.RegisterAddServer(s, srv)

	// 	logger.Log("addr", *grpcAddr)
	// 	errc <- s.Serve(ln)
	// }()

	// Enable - to connect to any gRPC service
	// Should be set to false in production
	if *debugAnyGRPCService {

		// HTTP transport for access to any internal service
		go func() {
			//logger := log.With(logger, "msg", "Debug Any Service", "transport", "HTTP")
			logger := log.With(logger, "tag", "#debughttp")
			h2 := addsvc.MakeDebugHTTPHandler(endpoints, tracer, logger)
			logger.Log("addr", *httpAnyServiceAddr, "tag", "#setup")
			errc <- http.ListenAndServe(*httpAnyServiceAddr, h2)
		}()

		// gRPC transport for access to any gRPC service.
		go func() {
			logger := log.With(logger, "msg", "Debug Any Service", "transport", "gRPC")

			ln, err := net.Listen("tcp", *gRPCAnyServiceAddr)
			if err != nil {
				errc <- err
				return
			}

			srv2 := addsvc.MakeAllServicesGRPCServer(endpoints, tracer, logger)
			s2 := grpc.NewServer()
			//pb.RegisterAddServer(s, srv)
			grpc_types.RegisterHelloServer(s2, srv2)
			grpc_types.RegisterWorldServer(s2, srv2)

			logger.Log("addr", *gRPCAnyServiceAddr)
			errc <- s2.Serve(ln)
		}()
	}

	// Run!
	logger.Log("exit", <-errc)
}
