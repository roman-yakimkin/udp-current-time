package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"udp-current-time/config"
)

type Server struct {
	cfg *config.Config
	log *log.Logger
}

func NewServer(cfg *config.Config, log *log.Logger) *Server {
	return &Server{
		cfg: cfg,
		log: log,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("resolve addr error: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen UDP error: %w", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	reqCh := make(chan *net.UDPAddr, 100)
	errCh := make(chan error, 1)

	wg := &sync.WaitGroup{}

	go func() {
		buf := make([]byte, 1024)
		for {
			_, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			select {
			case reqCh <- addr:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		<-ctx.Done()
		_ = conn.SetReadDeadline(time.Now())
	}()

	for {

		select {
		case addr := <-reqCh:
			wg.Add(1)
			go func(addr *net.UDPAddr) {
				defer wg.Done()

				currentTime := time.Now().Format(time.RFC3339)

				_, err = conn.WriteToUDP([]byte(currentTime), addr)
				if err != nil {
					log.Printf("failed to reply: %v", err)
				}
			}(addr)

		case err := <-errCh:
			if ctx.Err() != nil {
				wg.Wait()

				return ctx.Err()
			}

			return fmt.Errorf("reader failed: %w", err)

		case <-ctx.Done():
			wg.Wait()

			return ctx.Err()
		}
	}
}
