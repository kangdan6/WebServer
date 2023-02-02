package web

//Middleware 叫做函数式的责任链模式
// 函数式的洋葱模式
type Middleware func(next HandleFunc) HandleFunc


// AOP 方案在不同框架，不同语言里面都有不同的叫法
//比如Middleware, Handler, Chain, Filter, Filter-Chain, Interceptor, Wrapper


//这种叫 非函数式
//type MiddlewareV1 interface {
//	Invoke(next HandleFunc)HandleFunc
//}
//
//// 这种叫做拦截器
//type Interceptor interface {
//	Before(ctx *Context)
//	After(ctx *Context)
//	Surround(ctx *Context)
//}
//
//
//
//
// //还有下面这种变种
// type HandleFunc1 func(ctx *Context) (next bool)
// //和gin类似  集中式
// type Chain []HandleFunc1


// type ChainV1 struct {
// 	handlers []HandleFunc1
// }

// func (c ChainV1 ) Run(ctx *Context)  {
// 	for _, h := range c.handlers{
// 		next := h(ctx)
// 		//中断
// 		if !next{
// 			return
// 		}
// 	}
// }


//还有下面  一些业务分为各个步骤，每个步骤可以并发执行，每个步骤里又有小的步骤执行
//type HandleFuncV2 struct {
//	concurrent bool
//	handlers []*HandleFuncV2
//}
//
//func (h HandleFuncV2) Run(ctx *Context)  {
//	for _,hdl := range c.handlers{
//		h := hdl
//		if h.concurrent{
//			wg.Add(1)
//			go func(){
//				h.Run(ctx)
//				wg.Done()
//			}()
//		}
//	}
//}
//
//type Net struct {
//	handlers []HandleFuncV2
//}
//
//func (c Net) Run(ctx *Context)  {
//	var wg sync.WaitGroup
//	for _,hdl := range c.handlers{
//		h := hdl
//		//可以并发跑
//		if h.concurrent{
//			wg.Add(1)
//			go func(){
//				h.Run(ctx)
//				wg.Done()
//			}()
//		}else{
//			h.Run(ctx)
//		}
//	}
//	wg.Wait()
//}


