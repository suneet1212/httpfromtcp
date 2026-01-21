# HTTP From TCP

Following boot.dev course on building an http 1.1 server from scratch in GoLang. Only assuming that there exists a tcp library implemented.

Build the http parser step by step and use it to create a http server. Steps to build the server:

 <ol>
   <li>Parsing the Incoming Request
     <ol>
       <li>Read and save the request line</li>
       <li>Parse the headers</li>
       <li>Parse the body</li>
     </ol>
   </li>
   <li>Create a Response Writer
     <ol>
       <li>Write status line</li>
       <li>Write Headers</li>
       <li>Write Body</li>
       <li>Allow for chunked body</li>
       <li>Write the trailers</li>
     </ol>
   </li>
 </ol>

Now that the http protocol is ready, this can be used to create and host servers. 

To see it in action:
* Clone the repo and run it. This sets up a server running on port 42069 and waits for requests.
  ```
  cd httpfromtcp
  go run ./cmd/httpserver
  ```
* Open the following on your browser: <a href="http://localhost:42069/video">http://localhost:42069/video</a>
