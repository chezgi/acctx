package cli

import (
	"acctx/internal/output"
	"fmt"
)

func runVersion(a *app,args []string)error{data:=a.build;if a.json{return output.Write(a.streams.Out,true,output.Result{Command:"version",Status:"ok",Data:data})};return output.Write(a.streams.Out,false,output.Result{Command:"version",Status:"ok",Data:fmt.Sprintf("acctx %s (%s) %s",data.Version,data.Commit,data.BuiltAt)})}
