package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"ovs-switch/pkg/ovs"
)

type Node struct {
	Name          string `json:"name"`
	NodeIP        string `json:"nodeIP"`
	NeighborNodes []Node `json:"neighborNodes"`
}

// Script that takes two required arguments:
// the first one is the name in the cluster of the node where the script is running
// the second one is the path to the configuration file, in reference to the code.
func main() {

	configDir, nodeName, fileType, err := takeArguments()

	bridge := ovs.FromName("brtun")

	if err != nil {
		fmt.Println("Error with the arguments. Error:", err)
		return
	}

	nodes, err := readFile(configDir)

	switch fileType {
	case "topology":
		err = createTopology(bridge, nodes, nodeName)

	case "neighbors":
		err = connectToNeighbors(bridge, nodes[0])
	}

	if err != nil {
		fmt.Println("Vxlans not created: ", err)
		return
	}
}

func takeArguments() (string, string, string, error) {
	configDir := os.Args[len(os.Args)-1]

	nodeName := flag.String("node_name", "", "name of the node the script is executed in. Required.")

	fileType := flag.String("file_type", "topology", "type of filed passed as an argument. Can either be topology or neighbors. Default: topology.")

	flag.Parse()

	switch {
	case *nodeName == "":
		return "", "", "", errors.New("node name is not defined")
	case *fileType != "topology" || *fileType != "neighbors":
		return "", "", "", errors.New("file type not supported. Available types: 'topology' and 'neighbors'")
	case configDir == "":
		return "", "", "", errors.New("config directory is not defined")
	}

	return configDir, *nodeName, *fileType, nil
}

func createTopology(bridge ovs.Bridge, nodes []Node, nodeName string) error {

	// Search for the corresponding node in the configuration, according to the first passed parameter.
	// Once the node is found, create a bridge for every neighbour node defined.
	// The bridge is created with the nodeIp and neighborNodeIP and VNI. The VNI is generated in the l2sm-controller thats why its set to 'flow'.
	for _, node := range nodes {
		if node.Name == nodeName {
			//nodeIP := strings.TrimSpace(node.NodeIP)
			connectToNeighbors(bridge, node)
		}
	}
	return nil
}

func readFile(configDir string) ([]Node, error) {
	/// Read file and save in memory the JSON info
	data, err := ioutil.ReadFile(configDir)
	if err != nil {
		fmt.Println("No input file was found.", err)
		return nil, err
	}

	var nodes []Node
	err = json.Unmarshal(data, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil

}

func connectToNeighbors(bridge ovs.Bridge, node Node) error {
	for vxlanNumber, neighbor := range node.NeighborNodes {
		vxlanId := fmt.Sprintf("vxlan%d", vxlanNumber)
		err := bridge.CreateVxlan(ovs.Vxlan{VxlanId: vxlanId, LocalIp: node.NodeIP, RemoteIp: neighbor.NodeIP, UdpPort: "7000"})

		if err != nil {
			return fmt.Errorf("could not create vxlan between node %s and node %s", node.Name, neighbor)
		} else {
			fmt.Printf("Created vxlan between node %s and node %s.\n", node.Name, neighbor)
		}
	}
	return nil
}
