//--kind golang:default
//--main Hello

 // Hello function for the action
package main

 import "fmt"
 
 func Hello(obj map[string]interface{}) map[string]interface{} {
	 name, ok := obj["name"].(string)
	 if !ok {
		 name = "world"
	 }
	 fmt.Printf("name=%s\n", name)
	 msg := make(map[string]interface{})
	 msg["single-hello"] = "Hello, " + name + "!"
	 return msg
 }