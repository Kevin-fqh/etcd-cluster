package etcd

import (
	"bytes"
	"encoding/json"

	"io/ioutil"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

func (etcd_Node Etcd_Node) Listen_job() {

	//监听7070端口，等待连接，所有节点都运行
	ws := new(restful.WebService)
	ws.Route(ws.GET("/{node_name}/{node_ip}").To(etcd_Node.Get_parameter))
	restful.Add(ws)
	http.ListenAndServe(":7070", nil)
}

func (etcd_Node Etcd_Node) Get_parameter(request *restful.Request, response *restful.Response) {
	node_name := request.PathParameter("node_name")
	node_ip := request.PathParameter("node_ip")
	node_name = strings.Split(node_name, "=")[1]
	node_ip = strings.Split(node_ip, "=")[1]
	//执行add member 命令
	etcd_name, ectd_cluster, cluster_state := etcd_Node.Add_member(node_name, node_ip)
	//反馈结果
	cluster_env := Cluster_Info{etcd_name, ectd_cluster, cluster_state}
	message, _ := json.Marshal(cluster_env)
	response.Write(message)

}

func (etcd_Node Etcd_Node) Connect_to_cluster() {
	var url_buffer bytes.Buffer
	url_buffer.WriteString("http://")
	url_buffer.WriteString(etcd_Node.Leader_Ip)
	url_buffer.WriteString(":7070")
	url_buffer.WriteString("/node_name=")
	url_buffer.WriteString(etcd_Node.Node_Name)
	url_buffer.WriteString("/node_ip=")
	url_buffer.WriteString(etcd_Node.Node_Ip)
	url := url_buffer.String()
	ret, err := http.Get(url)
	//json格式
	if err != nil {
		glog.Infof("Error: %s", err)
	}
	defer ret.Body.Close()
	body, err := ioutil.ReadAll(ret.Body)
	if err != nil {
		glog.Infof("Error: %s", err)
	}
	var cluster_info Cluster_Info
	err = json.Unmarshal(body, &cluster_info)
	if err != nil {
		glog.Infof("Error: %s", err)
	}
	//拉起新节点的etcd服务
	state := strings.Split(cluster_info.Cluster_state, "=")
	clusters := strings.SplitAfterN(cluster_info.Ectd_cluster, "=", 2)
	state[1] = strings.Replace(state[1], "\"", "", -1)
	clusters[1] = strings.Replace(clusters[1], "\"", "", -1)
	etcd_Node.Start_etcd(state[1], clusters[1])

}
