package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

var podSandboxCommand = cli.Command{
	Name: "pod",
	Subcommands: []cli.Command{
		runPodSandboxCommand,
		stopPodSandboxCommand,
		removePodSandboxCommand,
		podSandboxStatusCommand,
		listPodSandboxCommand,
	},
}

var runPodSandboxCommand = cli.Command{
	Name:  "create",
	Usage: "create a pod",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "config.json",
			Usage: "the path of a pod sandbox config file",
		},
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "the name of the pod sandbox",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		// Test RuntimeServiceClient.RunPodSandbox
		err = RunPodSandbox(client, context.String("config"), context.String("name"))
		if err != nil {
			return fmt.Errorf("Creating the pod sandbox failed: %v", err)
		}
		return nil
	},
}

var stopPodSandboxCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a pod sandbox",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the pod sandbox",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = StopPodSandbox(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("stopping the pod sandbox failed: %v", err)
		}
		return nil
	},
}

var removePodSandboxCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a pod sandbox",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the pod sandbox",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = RemovePodSandbox(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("removing the pod sandbox failed: %v", err)
		}
		return nil
	},
}

var podSandboxStatusCommand = cli.Command{
	Name:  "status",
	Usage: "return the status of a pod",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "id of the pod",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = PodSandboxStatus(client, context.String("id"))
		if err != nil {
			return fmt.Errorf("getting the pod sandbox status failed: %v", err)
		}
		return nil
	},
}

var listPodSandboxCommand = cli.Command{
	Name:  "list",
	Usage: "list pod sandboxes",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "list only pod IDs",
		},
	},
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewRuntimeServiceClient(conn)

		err = ListPodSandboxes(client, context.Bool("quiet"))
		if err != nil {
			return fmt.Errorf("listing pod sandboxes failed: %v", err)
		}
		return nil
	},
}

// RunPodSandbox sends a RunPodSandboxRequest to the server, and parses
// the returned RunPodSandboxResponse.
func RunPodSandbox(client pb.RuntimeServiceClient, path string, name string) error {
	config, err := loadPodSandboxConfig(path)
	if err != nil {
		return err
	}

	// Override the name by the one specified through CLI
	if name != "" {
		config.Metadata.Name = &name
	}

	r, err := client.RunPodSandbox(context.Background(), &pb.RunPodSandboxRequest{Config: config})
	if err != nil {
		return err
	}
	fmt.Println(*r.PodSandboxId)
	return nil
}

// StopPodSandbox sends a StopPodSandboxRequest to the server, and parses
// the returned StopPodSandboxResponse.
func StopPodSandbox(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := client.StopPodSandbox(context.Background(), &pb.StopPodSandboxRequest{PodSandboxId: &ID})
	if err != nil {
		return err
	}
	fmt.Println(ID)
	return nil
}

// RemovePodSandbox sends a RemovePodSandboxRequest to the server, and parses
// the returned RemovePodSandboxResponse.
func RemovePodSandbox(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	_, err := client.RemovePodSandbox(context.Background(), &pb.RemovePodSandboxRequest{PodSandboxId: &ID})
	if err != nil {
		return err
	}
	fmt.Println(ID)
	return nil
}

// PodSandboxStatus sends a PodSandboxStatusRequest to the server, and parses
// the returned PodSandboxStatusResponse.
func PodSandboxStatus(client pb.RuntimeServiceClient, ID string) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	r, err := client.PodSandboxStatus(context.Background(), &pb.PodSandboxStatusRequest{PodSandboxId: &ID})
	if err != nil {
		return err
	}
	fmt.Printf("ID: %s\n", *r.Status.Id)
	if r.Status.Metadata != nil {
		if r.Status.Metadata.Name != nil {
			fmt.Printf("Name: %s\n", *r.Status.Metadata.Name)
		}
		if r.Status.Metadata.Uid != nil {
			fmt.Printf("UID: %s\n", *r.Status.Metadata.Uid)
		}
		if r.Status.Metadata.Namespace != nil {
			fmt.Printf("Namespace: %s\n", *r.Status.Metadata.Namespace)
		}
		if r.Status.Metadata.Attempt != nil {
			fmt.Printf("Attempt: %v\n", *r.Status.Metadata.Attempt)
		}
	}
	if r.Status.State != nil {
		fmt.Printf("Status: %s\n", r.Status.State)
	}
	if r.Status.CreatedAt != nil {
		ctm := time.Unix(*r.Status.CreatedAt, 0)
		fmt.Printf("Created: %v\n", ctm)
	}
	if r.Status.Linux != nil {
		fmt.Printf("Network namespace: %s\n", *r.Status.Linux.Namespaces.Network)
	}
	if r.Status.Network != nil {
		fmt.Printf("IP Address: %v\n", *r.Status.Network.Ip)
	}
	if r.Status.Labels != nil {
		fmt.Println("Labels:")
		for k, v := range r.Status.Labels {
			fmt.Printf("\t%s -> %s\n", k, v)
		}
	}
	if r.Status.Annotations != nil {
		fmt.Println("Annotations:")
		for k, v := range r.Status.Annotations {
			fmt.Printf("\t%s -> %s\n", k, v)
		}
	}
	return nil
}

// ListPodSandboxes sends a ListPodSandboxRequest to the server, and parses
// the returned ListPodSandboxResponse.
func ListPodSandboxes(client pb.RuntimeServiceClient, quiet bool) error {
	r, err := client.ListPodSandbox(context.Background(), &pb.ListPodSandboxRequest{})
	if err != nil {
		return err
	}
	for _, pod := range r.Items {
		if quiet {
			fmt.Println(*pod.Id)
			continue
		}
		fmt.Printf("ID: %s\n", *pod.Id)
		if pod.Metadata != nil {
			if pod.Metadata.Name != nil {
				fmt.Printf("Name: %s\n", *pod.Metadata.Name)
			}
			if pod.Metadata.Uid != nil {
				fmt.Printf("UID: %s\n", *pod.Metadata.Uid)
			}
			if pod.Metadata.Namespace != nil {
				fmt.Printf("Namespace: %s\n", *pod.Metadata.Namespace)
			}
			if pod.Metadata.Attempt != nil {
				fmt.Printf("Attempt: %v\n", *pod.Metadata.Attempt)
			}
		}
		fmt.Printf("Status: %s\n", pod.State)
		ctm := time.Unix(*pod.CreatedAt, 0)
		fmt.Printf("Created: %v\n", ctm)
		if pod.Labels != nil {
			fmt.Println("Labels:")
			for k, v := range pod.Labels {
				fmt.Printf("\t%s -> %s\n", k, v)
			}
		}
		if pod.Annotations != nil {
			fmt.Println("Annotations:")
			for k, v := range pod.Annotations {
				fmt.Printf("\t%s -> %s\n", k, v)
			}
		}
	}
	return nil
}
