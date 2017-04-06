package etcd

import (
	"bufio"
	"bytes"
	"exec"

	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
)

func New_Etcd_Node(is_leader string, etcd_leader_ip string, etcd_node_name string, etcd_node_ip string) Etcd_Node {
	//获取各种参数创建一个Etcd_Node对象
	glog.Infof("Info: is_leader: %s and etcd_leader_ip: %s", is_leader, etcd_leader_ip)
	hostname, _ := get_hostname()
	node_ip, _ := get_ip(hostname, etcd_node_ip)
	etcd_Node := Etcd_Node{
		Node_Name: etcd_node_name,
		Node_Ip:   node_ip,
		Is_Leader: is_leader,
		Leader_Ip: etcd_leader_ip,
	}
	return etcd_Node
}

func (etcd_Node Etcd_Node) Add_member(Node_Name string, Node_Ip string) (name string, cluster string, state string) {
	err := etcd_Node.Remove_member(Node_Name)
	if err != nil {
		glog.Infof("Error: %s", err)
		return
	}
	//执行add member 命令
	glog.Infof("excute etcdctl member add")
	var cmd_buffer bytes.Buffer
	cmd_buffer.WriteString("etcdctl member add ")
	cmd_buffer.WriteString(Node_Name)
	cmd_buffer.WriteString(" ")
	cmd_buffer.WriteString("http://")
	cmd_buffer.WriteString(Node_Ip)
	cmd_buffer.WriteString(":2380")
	cmd_line := cmd_buffer.String()
	result, err := exec.Exec_command(cmd_line)
	glog.Infof(result)
	if err != nil {
		glog.Infof("Error: %s", err)
		return
	}
	env := strings.Split(result, "\n")
	etcd_name := env[2]
	ectd_cluster := env[3]
	cluster_state := env[4]
	return etcd_name, ectd_cluster, cluster_state
}
func (etcd_Node Etcd_Node) Remove_member(Node_Name string) error {
	//leader节点在add member之前，先检测是否已经存在同名的etcd，若有，删除
	members, err := etcd_Node.get_member_list()
	if err != nil {
		glog.Infof("Error: %s", err)
		return err
	}
	if _, exist := members[Node_Name]; exist {
		//remove
		glog.Infof("Info: %s already exist,excute etcdctl member remove to delete it from etcd_cluster", Node_Name)
		cmd_line := "etcdctl member remove " + members[Node_Name]
		result, _ := exec.Exec_command(cmd_line)
		glog.Infof("Info: %s", result)
	}
	return nil
}
func (etcd_Node Etcd_Node) get_member_list() (map[string]string, error) {
	//返回Name的数组
	cmd_line := "etcdctl member list"
	members_results, err := exec.Exec_command(cmd_line)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(members_results, "\n")
	//8977d3f4a9aa1c55: name=bbb peerURLs=http://192.168.91.112:2380 clientURLs=http://192.168.91.112:2379 isLeader=false
	//value是ID key是name
	var members map[string]string
	members = make(map[string]string)
	for _, line := range lines {
		if len(line) <= 0 {
			break
		}
		members_temp := strings.Split(line, " ")
		members_temp[0] = strings.Replace(members_temp[0], ":", "", -1)
		members_temp[1] = strings.Replace(members_temp[1], "name=", "", -1)
		members[members_temp[1]] = members_temp[0]
	}

	return members, nil
}
func (etcd_Node Etcd_Node) Start_etcd(etcd_initial_cluster_state string, etcd_initial_cluster string) {
	var cmd_buffer bytes.Buffer
	cmd_buffer.WriteString("etcd -name ")
	cmd_buffer.WriteString(etcd_Node.Node_Name)
	cmd_buffer.WriteString(" ")
	cmd_buffer.WriteString("-data-dir /var/lib/etcd ")
	cmd_buffer.WriteString(" -listen-peer-urls http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2380")
	cmd_buffer.WriteString(" -initial-advertise-peer-urls http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2380")
	cmd_buffer.WriteString(" -listen-client-urls http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2379,http://127.0.0.1:2379")
	cmd_buffer.WriteString(" -advertise-client-urls http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2379")
	cmd_buffer.WriteString(" -initial-cluster ")
	change_peerURLs_flag := false
	if etcd_initial_cluster_state == "new" {
		cmd_buffer.WriteString(etcd_Node.Node_Name)
		cmd_buffer.WriteString("=http://")
		cmd_buffer.WriteString(etcd_Node.Node_Ip)
		cmd_buffer.WriteString(":2380")

		//如果/var/lib/etcd中已经有member文件夹,用old数据重建集群
		if _, err := os.Stat("/var/lib/etcd/member"); err == nil {
			glog.Infof("Info: /var/lib/etcd/member already exist,use the old data to rebuild etcd_cluster")
			//用old数据重建集群
			etcd_Node.rebuild_cluster()
			cmd_buffer.WriteString(" ")
			cmd_buffer.WriteString("-force-new-cluster")
			//设置标记 change_peerURLs_flag
			change_peerURLs_flag = true
		} else {
			cmd_buffer.WriteString(" -initial-cluster-state new")
		}
	} else {
		cmd_buffer.WriteString(etcd_initial_cluster)
		cmd_buffer.WriteString(" -initial-cluster-state existing")
		//新增的节点都清空 /var/lib/etcd
		os.RemoveAll("/var/lib/etcd")
	}
	cmd_buffer.WriteString(">> /var/log/etcd.log 2>&1 &")
	cmd_line := cmd_buffer.String()

	_, err := exec.Exec_command(cmd_line)
	glog.Infof("Info: %s", cmd_line)
	if err != nil {
		glog.Infof("Error: %s", err)
	}
	//针对重建的情况更改peerURLs
	if change_peerURLs_flag {
		//5s
		time.Sleep(5e9)
		err := etcd_Node.change_peerURLs()
		if err != nil {
			glog.Infof("Error: %s", err)
		}
	}
}
func (etcd_Node Etcd_Node) change_peerURLs() error {
	glog.Infof("Info: change_peerURLs")
	var cmd_buffer bytes.Buffer
	cmd_buffer.WriteString("curl http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2379/v2/members/")
	members, err := etcd_Node.get_member_list()
	if err != nil {
		glog.Infof("Error: %s", err)
		return err
	}
	ID := members[etcd_Node.Node_Name]
	cmd_buffer.WriteString(ID)
	cmd_buffer.WriteString(" -XPUT -H \"Content-Type:application/json\" -d '{\"peerURLs\":[\"http://")
	cmd_buffer.WriteString(etcd_Node.Node_Ip)
	cmd_buffer.WriteString(":2380\"]}'")
	cmd_line := cmd_buffer.String()
	_, err = exec.Exec_command(cmd_line)
	if err != nil {
		return err
	}
	return nil
}
func (etcd_Node Etcd_Node) rebuild_cluster() error {
	glog.Infof("Info: backup the old data")
	//备份,清空/var/lib/etcd ，然后把备份数据迁回来
	if _, err := os.Stat("/root/etcd_backup"); err != nil {
		err := os.MkdirAll("/root/etcd_backup", os.ModePerm)
		if err != nil {
			return err
		}
	}
	cmd_line := "etcdctl backup --data-dir /var/lib/etcd --backup-dir /root/etcd_backup"
	_, err := exec.Exec_command(cmd_line)
	if err != nil {
		return err
	}
	os.RemoveAll("/var/lib/etcd")
	cmd_l := "mv /root/etcd_backup/* /var/lib/etcd/"
	_, err = exec.Exec_command(cmd_l)
	if err != nil {
		return err
	}
	return nil
}

func get_hostname() (string, error) {
	cmd_line := "hostname"
	hostname, err := exec.Exec_command(cmd_line)
	if err != nil {
		return "", err
	}
	hostname = strings.Replace(hostname, "\n", "", -1)
	hostname = strings.Replace(hostname, " ", "", -1)
	return hostname, nil
}
func get_ip(hostname string, etcd_node_ip string) (string, error) {
	hosts_path := "/etc/hosts"
	hosts_file, err := os.Open(hosts_path)
	if err == nil {
		buff := bufio.NewReader(hosts_file)
		for {
			line, err := buff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			if strings.Contains(line, hostname) {
				//ubuntu中是个tab键  \t 分割。。。。 centos 是空格" "
				reip, _ := regexp.Compile(`(\d{1,3}\.){3}\d{1,3}`)
				node_ip := reip.FindAllString(line, -1)
				return node_ip[0], nil
			}
		}
	}
	if err != nil {
		return "error", err
	}
	//使用指定的ip（使用host模式启动容器的时候,/etc/hosts不存在对应条目）
	return etcd_node_ip, nil

}
