package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"gee_RPC/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// Server 服务端的实现
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(listen net.Listener) {
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServerConn(conn)
	}
}

func (s *Server) ServerConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	//f是用于处理对应编码类型的函数
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	//将连接conn交由对应的编码函数处理,获得实现了Codec接口的对象cc,将cc交给serverCodec进一步处理
	s.serverCodec(f(conn))
}

var invalidReq = struct{}{}

func (s *Server) serverCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Err = err
			s.sendResponse(cc, req.h, invalidReq, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h            *codec.Header // header of request
	args, replyv reflect.Value
}

func (s *Server) readRequestHeader(cc codec.Codec) (header *codec.Header, err error) {
	var h codec.Header
	if err = cc.ReadHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.args = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.args.Interface()); err != nil {
		log.Println("rpc server: read body error:", err)
	}
	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("rpc server: receive request:", req.h, req.args.Elem())
	fmt.Println()
	req.replyv = reflect.ValueOf(fmt.Sprintf("gee rpv resp : %d", req.h.Seq))
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func Accept(listen net.Listener) { DefaultServer.Accept(listen) }
