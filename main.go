// etcd_install project main.go
package main

import (
	"etcd"
	"flag"

	"github.com/golang/glog"
)

func main() {

	is_leader := flag.String("is_leader", "false", "defult value:false. true/false")
	etcd_leader_ip := flag.String("leader_ip", "127.0.0.1", "etcd cluster leader's ip")
	etcd_node_name := flag.String("node_name", "Node1", "the hostname")
	etcd_node_ip := flag.String("etcd_ip", "127.0.0.1", "the ip where is etcd run on. port is default!")
	flag.Parse()
	defer glog.Flush()

	etcd_Node := etcd.New_Etcd_Node(*is_leader, *etcd_leader_ip, *etcd_node_name, *etcd_node_ip)

	if etcd_Node.Is_Leader == "true" {
		//leader节点，拉起一个只有一个节点的etcd集群
		glog.Infof("Info: creating a new etcd_cluster.If the old data is provided,it will create the cluster base on the old data.")
		etcd_Node.Start_etcd("new", "")
	} else {
		glog.Infof("Info: join to an existing etcd_cluster accroding the leader_ip")
		etcd_Node.Connect_to_cluster()
		//根据leder_ip加入已有的etcd集群
	}
	////监听7070端口，等待新节点连接，所有节点都运行
	etcd_Node.Listen_job()
	//后续记录维持集群的成员信息、节点的健康信息（自身的etcd进程是否正常），及作出后续操作
}
