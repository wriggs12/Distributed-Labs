package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"


type Coordinator struct {
	FilesToDo []string
	FilesInProg []string
	FilesDone []string
	NReduce int
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) RequestTask(args *ReqTaskArgs, reply *ReqTaskReply) error {
	if len(c.FilesToDo) != 0 {
		reply.FileName = c.FilesToDo[0]
		
		c.FilesInProg = append(c.FilesInProg, c.FilesToDo[0])
		c.FilesToDo = c.FilesToDo[1:]
	} else {
		reply.FileName = ""
	}

	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := len(c.FilesToDo) == 0 && len(c.FilesInProg) == 0

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	c.FilesToDo = files
	c.NReduce = nReduce
	
	c.server()
	return &c
}
