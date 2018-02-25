## How to run this app on a local server.

### Requirements:
* Node.js installed.
* Golang installed.
* MySQL Server installed and running.

### Follow these steps to run the app:
* #### Using CLI (i.e. command line), download and access the repository on your computer:
    `git clone https://github.com/bulllittt1/react-my-tree.git -b server-api && cd react-my-tree`
* #### In the project directory:
    * If you have node packages installed globally on your computer, just link to them: <br>
      `npm link`
    * If not, install node packages into the project directory: <br>
      `npm install`
 * #### Set up GOPATH:
      `export PATH=$PATH:$(go env GOPATH)/bin`
 * #### Download needed Go package:
      `go get github.com/rs/cors`
 * #### Run the backend:
    `go run server.go`
    * Console will ask $username and $password to access your MySQL server!
 * #### Open another console window in the same directory again (~/react-my-tree) and run the frontend; react-app in the development mode:
    `npm start`
* #### Open [http://localhost:3000](http://localhost:3000) in the browser.

The page will reload if you make edits.<br>
You will also see any errors and messages in the browser DevTools console.
