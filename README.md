# etch-cluster
a too which helps to deploy etcd cluster with the registered system


# usage
`go build -o etcd_install`

get the command etcd_install and then use the Dockerfile to build the image

get the master contaier run on the ip of 172.17.0.2 

`docker run -itd  -e is_leader=true -e node_name=aaa 98c2ef0aa9bb`

and then other node joins the master

`docker run -itd  -e is_leader=false -e leader_ip=172.17.0.2  -e node_name=bbb 98c2ef0aa9bb`

you can also specify the volume

`docker run -itd  -e is_leader=true -e node_name=aaa -v /var/lib/etcd:/var/lib/etcd 98c2ef0aa9bb`

you can also use the host network model ,specify the ip 

`docker run -itd --net=host -e is_leader=true -e node_name=aaa -e etcd_ip=192.168.56.101 4570ffde4e14`


