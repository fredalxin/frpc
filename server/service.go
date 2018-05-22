package server

import (
	"context"
	"errors"
	"fmt"
	"frpc/log"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

type methodType struct {
	sync.Mutex
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
	numCalls  uint
}

type service struct {
	name   string
	rcvr   reflect.Value
	typ    reflect.Type
	method map[string]*methodType
}

func (s *Server) Register(rcvr interface{}) (*Server, error) {
	name := reflect.Indirect(reflect.ValueOf(rcvr)).Type().Name()
	return s.RegisterWithMeta(rcvr, name, "")
}

func (s *Server) RegisterWithName(rcvr interface{}, name string) (*Server, error) {
	return s.RegisterWithMeta(rcvr, name, "")
}

func (s *Server) RegisterWithMeta(rcvr interface{}, name string, metadata string) (*Server, error) {
	if s.registryServer.registry != nil {
		s.registryServer.registry.Register(name, rcvr, metadata)
	}
	//todo controllerServer
	s.controllerServer.Register(name, rcvr, metadata)

	return s, s.register(rcvr, name, true)
}

func (s *Server) register(rcvr interface{}, name string, useName bool) error {
	s.serviceMapMu.Lock()
	defer s.serviceMapMu.Unlock()
	if s.serviceMap == nil {
		s.serviceMap = make(map[string]*service)
	}
	service := new(service)
	service.typ = reflect.TypeOf(rcvr)
	service.rcvr = reflect.ValueOf(rcvr)
	//deal name
	sname := name
	if !useName {
		sname = reflect.Indirect(service.rcvr).Type().Name()
	}
	if sname == "" {
		errorStr := "frpc.Register: no service name for type " + service.typ.String()
		log.Error(errorStr)
		return errors.New(errorStr)
	}
	if !useName && !isExported(sname) {
		errorStr := "frpc.Register: type " + sname + " is not exported"
		log.Error(errorStr)
		return errors.New(errorStr)
	}
	service.name = sname

	//deal method
	service.method = suitableMethods(service.typ, true)
	if len(service.method) == 0 {
		errorStr := "frpc.Register: type " + sname + " has no exported methods of suitable type"
		log.Error(errorStr)
		return errors.New(errorStr)
	}
	s.serviceMap[service.name] = service
	return nil
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

//suitableMethods returns suitable Rpc methods of typ, it will report
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if method.PkgPath != "" {
			continue
		}
		//methods needs four ins
		if mtype.NumIn() != 4 {
			if reportErr {
				log.Info("method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		//first in must be context.Context
		ctxType := mtype.In(1)
		if !ctxType.Implements(typeOfContext) {
			if reportErr {
				log.Info("method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		//seconds need not be a pointer  why?
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Info(mname, "argument type not exported:", argType)
			}
			continue
		}
		// Third must be a pointer.
		replyType := mtype.In(3)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				log.Info("method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				log.Info("method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				log.Info("method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				log.Info("method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &methodType{method: method, argType: argType, replyType: replyType}
	}
	return methods
}

func (s *service) call(ctx context.Context, mtype *methodType, argv, replyv reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error: %v", r)
		}
	}()
	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(ctx), argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	if errInter != nil {
		return errInter.(error)
	}

	return nil
}
