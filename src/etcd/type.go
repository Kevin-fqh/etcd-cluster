package etcd

type Etcd_Node struct {
	Node_Name string
	Node_Ip   string
	Is_Leader string
	Leader_Ip string
	Status    string
}
type Cluster_Info struct {
	Etcd_name, Ectd_cluster, Cluster_state string
}
