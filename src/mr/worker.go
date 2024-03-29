package mr

import "fmt"
import "log"
import "os"
import "io"
import "sort"
import "net/rpc"
import "hash/fnv"


//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	
	reply := ReqeustTask()
	FileName := reply.FileName
	OutNumber := reply.OutNumber

	for FileName != "" {
		DoWork(FileName, OutNumber, mapf, reducef)
		log.Printf("Completed File: %s\n", FileName)

		reply = ReqeustTask()
		FileName = reply.FileName
		OutNumber = reply.OutNumber
	}

	log.Printf("No Task Recieved, Exiting ...\n")

}

func DoWork(FileName string, OutNumber int,
	mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	file, err := os.Open(FileName)
	if err != nil {
		log.Fatalf("cannot open %v", FileName)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", FileName)
	}
	file.Close()
	kva := mapf(FileName, string(content))

	sort.Sort(ByKey(kva))

	oname := fmt.Sprintf("mr-out-%d", OutNumber)
	ofile, _ := os.Create(oname)

	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[i].Value)
		}
		output := reducef(kva[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)
	
		i = j
	}

	ofile.Close()
}

func ReqeustTask() ReqTaskReply {
	args := ReqTaskArgs{}
	reply := ReqTaskReply{}

	ok := call("Coordinator.RequestTask", &args, &reply)
	if ok {
		return reply
	} else {
		return reply
	}
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
