package web

import (
	"container/list"
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 注册路由。
// method 是 HTTP 方法
// path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
func (r *router) addRoute(method string, path string, handler HandleFunc, mdls ...Middleware) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	//获得树的根节点
	root, ok := r.trees[method]
	if !ok {
		// 创建根节点
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handler
		root.mdls = mdls //增加middleware
		return
	}

	//分割
	seg := strings.Split(path[1:], "/")
	for _, s := range seg {
		if s == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(s)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
	root.mdls = mdls //增加middleware

}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
// 同时注册/user/*和/user/*/home时，以/home开头的路由，按最长的规则匹配/user/*/home，否则匹配到/user/*,这种情况/user/123/home/456匹配到/user/*
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		root.route = "/"
		return &matchInfo{
			n:    root,
			mdls: root.mdls,
		}, true
	}
	path = strings.Trim(path, "/")
	segs := strings.Split(path, "/")
	mi := &matchInfo{}
	//如果匹配到*提前记录  这样就不用回溯了
	var mi_n *node
	cur := root
	for _, seg := range segs {
		cur, ok = cur.childof(seg)
		if !ok {
			if mi_n != nil {
				break
			} else {
				return nil, false
			}

		}
		if cur.typ == nodeTypeReg || cur.typ == nodeTypeParam {
			mi.addValue(cur.paramName, seg)
		}


		//记录中间匹配上的，避免回溯
		if cur != nil && cur.handler != nil{
			mi_n = cur
		}

		////必做题1，如果命中的是通配符路由
		//if cur.path == "*" && cur.handler != nil {
		//	mi_n = cur
		//}
	}
	if cur != nil {
		mi.n = cur
	} else {
		mi.n = mi_n
	}

	mi.n.route = path
	mi.mdls = r.findMdls(root, segs)

	return mi, true
}

// 层序遍历查找middleware
func (r *router) findMdls(root *node, segs []string) []Middleware {
	//  root.route = "/" 要先排除 不能入st
	mdls := []Middleware{}
	st := list.New()
	// 通配符-》正则路由 -》参数路由-》静态路由
	if root.starChild != nil {
		st.PushBack(root.starChild)
	}

	if root.regChild != nil{
		st.PushBack(root.regChild)
	}

	if root.paramChild != nil{
		st.PushBack(root.paramChild)
	}

	if root.children != nil {
		for _, staticNode := range root.children {
			st.PushBack(staticNode)
		}
	}
	if root.path == "/" && root.mdls != nil {
		mdls = append(mdls, root.mdls...)
	}

	layerIndex := 0
	for st.Len() > 0 && layerIndex < len(segs) {
		length := st.Len()
		path := segs[layerIndex]
		for i := 0; i < length; i++ {
			tmpnode := st.Remove(st.Front()).(*node)
			// 通配符-》正则路由 -》参数路由-》静态路由
			if tmpnode.typ == nodeTypeAny && tmpnode.mdls != nil {
				mdls = append(mdls, tmpnode.mdls...)
			} else if tmpnode.typ == nodeTypeReg && tmpnode.regExpr.Match([]byte(path))&& tmpnode.mdls != nil{
				mdls = append(mdls, tmpnode.mdls...)
			}else if tmpnode.typ == nodeTypeParam && tmpnode.mdls != nil{
				mdls = append(mdls, tmpnode.mdls...)
			}else if tmpnode.typ == nodeTypeStatic && tmpnode.path == path && tmpnode.mdls != nil {
				mdls = append(mdls, tmpnode.mdls...)
			}

			if tmpnode.starChild != nil {
				st.PushBack(tmpnode.starChild)
			}

			if tmpnode.regChild != nil{
				st.PushBack(tmpnode.regChild)
			}

			if tmpnode.paramChild != nil{
				st.PushBack(tmpnode.paramChild)
			}


			if tmpnode.children != nil {
				for _, staticNode := range tmpnode.children {
					st.PushBack(staticNode)
				}
			}
		}
		layerIndex++

	}
	return mdls
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 正则匹配，形式 :param_name(reg_expr)
// 3. 路径参数匹配：形式 :param_name
// 4. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	//匹配的完整路径
	route string

	// 路径
	path string
	// 子节点的 path => node
	children map[string]*node
	// 通配符 * 表达的节点，任意匹配
	starChild *node
	//参数匹配
	paramChild *node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp

	//middleware
	mdls []Middleware
}

// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由、正则路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由、正则路由和参数路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path, typ: nodeTypeAny}
		}
		return n.starChild

	}
	// 以 : 开头，我们认为是参数路由
	if path[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由、正则路由和参数路由 [%s]", path))
		}

		index1 := strings.Index(path, "(")
		index2 := strings.Index(path, ")")
		//正则路由
		if index1 != -1 && index2 != -1 {
			if n.regChild != nil {
				if n.regChild.path != path {
					panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
				}
			} else {
				reg, err := regexp.Compile(path[index1+1 : index2])
				if err != nil {
					panic(fmt.Sprintf("web: 正则表达错误，%s", path[index1+1:index2]))
				}
				n.regChild = &node{path: path, typ: nodeTypeReg, paramName: path[1:index1], regExpr: reg}
			}
			return n.regChild
		}
		// 以 : 开头，我们认为是参数路由
		if n.paramChild != nil {
			if n.paramChild.path != path {
				panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
			}
		} else {
			n.paramChild = &node{path: path, typ: nodeTypeParam, paramName: path[1:]}
		}
		return n.paramChild
	}
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path, typ: nodeTypeStatic}
		n.children[path] = child
	}
	return child
}

// childof 优先考虑静态匹配，匹配不上再考虑通配符匹配
// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否命中
func (n *node) childof(path string) (*node, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true
		}
		if n.regChild != nil {
			return n.regChild, true
		}
		return n.starChild, n.starChild != nil
	}
	child, ok := n.children[path] //注意这里不要用n同名，会被赋值nil
	// 针对注册了路由 /user/*/abc 那么遍历到user查找*时
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true
		}
		if n.regChild != nil {
			// 判断正则是否匹配
			if n.regChild.regExpr.Match([]byte(path)) {
				return n.regChild, true
			}
		}
		return n.starChild, n.starChild != nil
	}
	return child, ok
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
	mdls       []Middleware
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		m.pathParams = make(map[string]string)
	}
	m.pathParams[key] = value
}
