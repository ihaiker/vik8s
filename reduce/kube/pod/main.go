package pod

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod/container"
	"github.com/ihaiker/vik8s/reduce/kube/pod/volumes"
	"github.com/ihaiker/vik8s/reduce/plugins"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var handlers = map[string]func(directive *config.Directive, spec *v1.PodSpec){
	"container": container.ContainerParse,

	"init-container": container.InitContainerParse, "initContainer": container.InitContainerParse,

	"hosts":  HostAliasesParse,
	"volume": volumes.VolumeParse, "volumes": volumes.VolumesParse,
	"affinity": AffinityParse,
}

func PodSpecParse(directive *config.Directive, podSpec *v1.PodSpec) {
	for _, item := range directive.Body {
		if handler, has := handlers[item.Name]; has {
			handler(item, podSpec)
		} else {
			refs.UnmarshalItem(podSpec, item)
		}
	}
}

func parse(version, prefix string, directive *config.Directive) metav1.Object {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(pod, directive)
	asserts.AutoLabels(pod, prefix)
	PodSpecParse(directive, &pod.Spec)
	return pod
}

var Pod = plugins.ReduceHandler{Names: []string{"pod", "Pod"}, Handler: parse, Demo: `
pod test [label=value ...] {
    label labelN valueN;
    labels {
        label1 value1;
        label2 value2;
    }
    
    hostNetwork true;

    hosts 172.16.100.12 api.vik8s.io; 
    hosts {
        172.16.100.12 api.vik8s.io;
        172.16.100.13 api.vik8s.io api2.vik8s.io;
    }

    affinity pod {
        preferred weight:1 {
            labels {
                key-w2-1 ab;
                key-w2-2 ab;
            }
            expr {
                key-w1-1 In ab;
                key-w1-2 In ab;
            }
            namespaces n1 n2 n3;
        }
        required key2:1 {
            labels {
                key-w2-1 ab;
                key-w2-2 ab;
            }
            expr {
                key-w1-1 In ab;
                key-w1-2 In ab;
            }
            namespaces n1 n2 n3;
        }
        required key3:1 {
            labels {
                key-w2-1 ab;
                key-w2-2 ab;
            }
            expr {
                key-w1-1 In ab;
                key-w1-2 In ab;
            }
            namespaces n1 n2 n3;
        }
    }
    container continarName image [imagePullSecret] {
        args arg1 arg2 ... argN;
        command cmd1 cmd2 ... cmdN;
        resources {
            requests memory=100Mi cpu=100m;
            limit memory=100Mi cpu=100m;
        }
        envs {
            ENV_VALUE           test1123;
            ENV_CONFIGMAP       configMap   name1 key1;
            ENV_CONFIG          config      name1 key1;
            ENV_FIELD           field       spec.nodeName;
            ENV_RESOURCE_MEN    resource    container_name  requests.memory;
            ENV_RESOURCE_CPU    res         container_name  requests.cpu;
        }
        env ENV_SECRET          secret name12 k123;

        envFrom configmap:config-name [prefix];
        envFrom secret:secret-config-name [prefix];
    
        port [name0] [[hostIP:]hostPort:]containerPort/protocol;
        ports {
            [name1] [[hostIP:]hostPort:]containerPort/protocol;
            [name2] [[hostIP:]hostPort:]containerPort/protocol;
        }

        mount from:volume-name /data[:subPath];
        mount from:from-volume-name /mountPath {
            mountPropagation Bidirectional;
            subPath in;
            subPathExpr in.*;
            readOnly true;
        }

        mount empty:emtpy-volue-data /data/in;
        mount hostPath:etcd-config-data /data/etcd:/var/lib/etcd;
        mount hostPath:localtime /etc/localtime:/etc/localtime File;
        
        mount secret:secret-volume1 /data/secret;
    
        mount configMap:configmap-volume1 /data/config/sub:sub;
        mount configmap:volume-name:config-name /etc/nginx;
        mount configmap:volume-name:config-name /etc/nginx/nginx.conf:nginx.conf {
            #item.key:item.mode item.path;
            nginx.conf:0655 [nginx.conf];
        }

        mount glusterfs:gluster-mysql-pvc /data[:subPath] {
            endpoints enddd;
            path mysql;
        }

        mount pvc:volume-name /data;
        mount pvc:volume-name:pvc-name /data;
        mount pvc:volume-name:pvc-name /data[:subPath] {
            readOnly true;
        }
    }

    volume empty:empty-volume1;
    volume empty:empty-volume2 {
        medium "medium";
        sizeLimit "sizeLimit";
    }

    volume hostpath:hostpath-docker-volume1 /data/docker;
    volume hostpath:hostpath-volume2 /etc/hosts File;

    volume configmap:config-volume-1;
    volume configmap:config-volume-2:0655;
    volume configmap:config-volume-3 configmap-volume-config;
    volume configmap:config-volume-4 configmap-volume-config{
        k1 v1;
    }

    volumes {
        configmap:config-volume4 ;
        configmap:config-volume5 configmap-volume-config;
        configmap:config-volume6 configmap-volume-config {
            k1 v1;
        }
        configmap:config-volume7:0655 configmap-volume-config;
    }

    volume pvc:mysql-pvc;
    volume pvc:mysql-pvc:mysql-pvc-config2;
    volume pvc:mysql-pvc:mysql-pvc-config {
        readOnly true;
    }

    volume secret:config-volume1;
    volume secret:config-volume:secret-config;

    volume secret:config-volume1:secret-config-volume1 {
        defaultModule 0655;
        items {
            nginx.conf:0655 nginx.config;
        }
    }
    volume secret:config-volume1:secret-config-volume1 {
        defaultModule 0655;
        nginx.conf:0655 nginx.config;
    }

    volume cephfs:ceph-mysql-pvc {
        monitors monitorPath1 [monitorPath2 ...];
        secretRef {
            name b123;
        }
        secretFile a1;
    }

    volume glusterfs:glusetr-mysql {
        endpoints http://10.24.0.2:24007;
        path /data/mysql;
        readOnly false;
    }

    volume ceph:ceph-mysql {
        monitors name1 name2;
        path  /data/mysql;
        secretFile sec123;
        secretRef name=nam123;
    }
}
`}
