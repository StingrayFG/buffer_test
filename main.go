package main

import (
  "os"
  "fmt"
  "time"
  "log"

  "net/http"

  "strings"
  "strconv"

  "github.com/joho/godotenv"
)

// Variables to save environment variables' values.
var goBufferDestinationUrl string 
var goBufferQueueCheckPeriod time.Duration
var goBufferMaxErrorsCount int

//
var factsQueue []string // used to temporarily store the formdata contents before sending them to the destination
var factsAuthHeaderQueue []string // used to temporarily store authorization tokens
var isSending bool = false // shows whether facts are being sent to the destination at the moment


func addFactToQueue(w http.ResponseWriter, r *http.Request) { // used to save the formdata sent by client into the factsQueue
  err := r.ParseForm()
  if err != nil {
    log.Fatalln(err)
  }
  fmt.Println("add " + r.Form.Encode())

  factsQueue = append(factsQueue, r.Form.Encode()) // parse the formdata into string
  factsAuthHeaderQueue = append(factsAuthHeaderQueue, r.Header.Get("Authorization"))
}


func sendFactFromQueue(fact string, authHeader string) int { // used to send the first fact from the factsQueue to the destination
  fmt.Println("sent " + fact)
  client := &http.Client{}
  
  req, err := http.NewRequest("POST", goBufferDestinationUrl, strings.NewReader(fact)) // assemble post request
  if err != nil {
    log.Fatalln(err)
  }
  req.Header.Add("Content-Type", "application/x-www-form-urlencoded") 
  req.Header.Add("Authorization", authHeader) 

  
  res, err := client.Do(req) // send post request
  if err != nil {
    log.Fatalln(err)
  }
  //fmt.Println(res)
  defer res.Body.Close() // close the response body to mitigate errors
  return res.StatusCode
}


func beginSendingFromQueue() { // used to send facts from the factsQueue sequentially, one by one
  errorsCount := 0

  for (len(factsQueue) > 0) { // while loop to iterate over the factsQueue
    if (errorsCount >= goBufferMaxErrorsCount) { // limit POST request attempts to the destination endpoint to the specified value
      // if the limit gets exceeded, stop sending facts to the destination until the next checkQueue()'s tick
      errorsCount = 0
      isSending = false
      break
    }

    status := sendFactFromQueue(factsQueue[0], factsAuthHeaderQueue[0]) // try sending the first fact from the factsQueue
    fmt.Println("response " + strconv.Itoa(status))
    if(((status >= 200) && (status < 300)) || (status == 500)) { 
      /* ### CONDITION (status == 500) IS PRESENT ABOVE ONLY DUE TO THE EXAMPLE SERVER SENDING HTTP CODE 500 RESPONSES TO SUCCESSFUL POST REQUESTS. 
      IT SHALL BE REMOVED UPON FIGURING OUT THE REASON FOR SUCH BEHAVIOR ### */
      factsQueue = factsQueue[1:] // if suceeded, remove the sent fact from the factsQueue
      factsAuthHeaderQueue = factsAuthHeaderQueue[1:] // if suceeded, remove the sent fact's auth header from the factsAuthHeaderQueue
      errorsCount = 0
    } else { // if request fails
      errorsCount += 1
    }
  }

  isSending = false // set it to false as the sign of queue being empty
}


func checkQueue() { // used to check whether the factsQueue contains any facts
  ticker := time.NewTicker(time.Second * goBufferQueueCheckPeriod) // Create ticker to handle queue checking periods
	tickerChan := make(chan bool)	// Create channel using make

	for {
    select {
      case <-tickerChan:
        return
      case <-ticker.C: // on the ticker's tick
        if ((len(factsQueue) > 0) && (isSending == false)) { // if true, begin sending facts to destination
          // 'isSending == false' condition is used to prevent multiple simultaneous calls of beginSendingFromQueue()
          isSending = true
          beginSendingFromQueue()
        }
    }  
  }
}


func readEnvironmentVariables() { // used to save environment variables' values into global variables
  // Such approach prevents unnecessary os.Getenv() calls each time an environment variable's value is needed
  goBufferDestinationUrl = os.Getenv("GOBUFFER_DESTINATION_URL")
  val_qcp, err := strconv.Atoi(os.Getenv("GOBUFFER_QUEUE_CHECK_PERIOD")) 
  if err != nil {
    panic(err)
	} else {
    goBufferQueueCheckPeriod = time.Duration(val_qcp) // Allows saving the resulting value into the global variable and not the local one
    // It needs to be parsed to time.Duration type in order to be used to set the NewTicker object's time interval in checkQueue()
  }
  val_mec, err := strconv.Atoi(os.Getenv("GOBUFFER_QUEUE_CHECK_PERIOD")) 
  if err != nil {
    panic(err)
	} else {
    goBufferMaxErrorsCount = val_mec // Allows saving the resulting value into the global variable and not the local one
  }
}


func main() {   
  env := os.Getenv("GOBUFFER_PORT") /* Will get loaded before reading the .env file 
  only when this app is running as a docker container with environment variables already specified.
  Otherwise, it will return an empty string and will lead to the .env file being read */
  if (env == "") {
    err_env := godotenv.Load() // load environment variables
    if err_env != nil {
      log.Fatalln(err_env)
    }
  }
  readEnvironmentVariables() // save environment variables' values into global variables 

  go checkQueue() // invoke it in a goroutine, because checkQueue()'s needs to be run asynchronously due to it's ticker

  http.HandleFunc("/fact", addFactToQueue) // set endpoint's route
  err_las := http.ListenAndServe(":" + os.Getenv("GOBUFFER_PORT"), nil) // start on the given port
  if err_las != nil {
    log.Fatalln(err_las)
    return
  }
}