package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type MyService struct {
	UnimplementedAdminServer
	UnimplementedBizServer
	acl        map[string][]string
	addr       string
	eventsChan map[string]chan *Event
	stat       *statStore
	ctx        context.Context
}

type statStore struct {
	sync.Mutex
	stats map[string]*Stat
}

func NewMyService(addr, acl string) (*MyService, error) {
	parsedAcl := make(map[string][]string)
	err := json.Unmarshal([]byte(acl), &parsedAcl)
	if err != nil {
		return nil, fmt.Errorf("init service: %w", err)
	}

	service := &MyService{
		acl:        parsedAcl,
		addr:       addr,
		eventsChan: make(map[string]chan *Event),
		stat:       &statStore{stats: make(map[string]*Stat)}}

	return service, nil
}

func (s *MyService) Start(ctx context.Context) error {
	serviceCtx, cancel := context.WithCancel(ctx)
	s.ctx = serviceCtx

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	RegisterAdminServer(grpcServer, s)
	RegisterBizServer(grpcServer, s)

	go grpcServer.Serve(lis)

	<-ctx.Done()
	cancel()
	grpcServer.GracefulStop()

	return nil
}

func (s *MyService) isMethodAllowed(consumer, methodName string) bool {
	allowed := false
	for _, p := range s.acl[consumer] {
		if matched, _ := regexp.MatchString(p, methodName); matched {
			allowed = true
			break
		}
	}
	return allowed
}

func (s *MyService) initStatStore(requestID string) {
	s.stat.Lock()
	s.stat.stats[requestID] = &Stat{ByMethod: make(map[string]uint64), ByConsumer: make(map[string]uint64)}
	s.stat.Unlock()
}

func (s *MyService) deleteStatStore(requestID string) {
	s.stat.Lock()
	delete(s.stat.stats, requestID)
	s.stat.Unlock()
}

func (s *MyService) logEvent(requestID string, ts int64, consumer, method, host string) {
	event := &Event{Timestamp: ts, Consumer: consumer, Method: method, Host: host}

	for id, lis := range s.eventsChan {
		if id != requestID {
			lis <- event
		}
	}
	s.stat.Lock()
	for id := range s.stat.stats {
		if id != requestID {
			s.stat.stats[id].ByConsumer[consumer] += 1
			s.stat.stats[id].ByMethod[method] += 1
		}
	}
	s.stat.Unlock()
}

func fetchConsumerAndHost(ctx context.Context) (string, string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ""
	}

	consumer := fetchFirstMDValue(md, "consumer")
	fullHost := fetchFirstMDValue(md, ":authority")
	host := strings.Split(fullHost, ":")[0] + ":"

	return consumer, host
}

func fetchFirstMDValue(md metadata.MD, key string) string {
	var value string
	values := md.Get(key)
	if len(values) != 0 {
		value = values[0]
	}

	return value
}

// Admin service

func (s *MyService) Logging(_ *Nothing, stream grpc.ServerStreamingServer[Event]) error {
	requestID := uuid.NewString()
	consumer, host := fetchConsumerAndHost(stream.Context())

	go s.logEvent(requestID, time.Now().Unix(), consumer, Admin_Logging_FullMethodName, host)

	if !s.isMethodAllowed(consumer, Admin_Logging_FullMethodName) {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	s.eventsChan[requestID] = make(chan *Event, 5)

	for {
		select {
		case event := <-s.eventsChan[requestID]:
			if err := stream.Send(event); err != nil {
				return status.Errorf(codes.Internal, "Send error: %s", err)
			}
		case <-s.ctx.Done():
			delete(s.eventsChan, requestID)
			return nil
		}
	}
}

func (s *MyService) Statistics(interval *StatInterval, stream grpc.ServerStreamingServer[Stat]) error {
	requestID := uuid.NewString()

	consumer, host := fetchConsumerAndHost(stream.Context())

	go s.logEvent(requestID, time.Now().Unix(), consumer, Admin_Statistics_FullMethodName, host)

	if !s.isMethodAllowed(consumer, Admin_Statistics_FullMethodName) {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	s.initStatStore(requestID)

	for {
		select {
		case <-time.After(time.Duration(interval.GetIntervalSeconds()) * time.Second):
			stat := s.stat.stats[requestID]
			if err := stream.Send(stat); err != nil {
				return status.Errorf(codes.Internal, "Send error: %s", err)
			}

			s.initStatStore(requestID)
		case <-s.ctx.Done():
			s.deleteStatStore(requestID)
			return nil
		}
	}
}

// Biz service

func (s *MyService) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	requestID := uuid.NewString()

	consumer, host := fetchConsumerAndHost(ctx)

	go s.logEvent(requestID, time.Now().Unix(), consumer, Biz_Add_FullMethodName, host)

	if !s.isMethodAllowed(consumer, Biz_Add_FullMethodName) {
		return &Nothing{}, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	return &Nothing{}, nil
}

func (s *MyService) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	requestID := uuid.NewString()

	consumer, host := fetchConsumerAndHost(ctx)

	go s.logEvent(requestID, time.Now().Unix(), consumer, Biz_Check_FullMethodName, host)

	if !s.isMethodAllowed(consumer, Biz_Check_FullMethodName) {
		return &Nothing{}, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	return &Nothing{}, nil
}

func (s *MyService) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	requestID := uuid.NewString()

	consumer, host := fetchConsumerAndHost(ctx)

	go s.logEvent(requestID, time.Now().Unix(), consumer, Biz_Test_FullMethodName, host)

	if !s.isMethodAllowed(consumer, Biz_Test_FullMethodName) {
		return &Nothing{}, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	return &Nothing{}, nil
}

func StartMyMicroservice(ctx context.Context, addr string, ACLData string) error {
	s, err := NewMyService(addr, ACLData)
	if err != nil {
		return err
	}

	go s.Start(ctx)

	return nil
}
