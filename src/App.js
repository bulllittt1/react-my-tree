import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

// Sidebar receives user input("title" and "avatar") for a new Node,
// appears when "+" button is clicked on a Node.
class Sidebar extends Component {
  constructor(props) {
    super(props);
    this.state = {
      currentFile: null,
      currentTitle: ''
    };
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleSubmit() {
      const file = this.state.currentFile;
      const title = this.state.currentTitle;
      this.props.onSubmit(file, title);
  }

  handleInputChange(input) {
      this.setState({
        currentFile: input.target.files[0]
      });
  }

  handleTitleChange(e) {
      this.setState({
        currentTitle: e.target.value
      });
  }

  render() {
    return (
      <div className="Sidebar">
        <h3>Create Node </h3>
        <label>
          Upload image:
        </label>
        <input
          type="file"
          accept="image/*"
          onChange = {(input) => this.handleInputChange(input)}
        />
        <input
          type="text"
          placeholder='Enter a title'
          autoFocus
          onChange = {(e) => this.handleTitleChange(e)}
        />
        <button id='btn-createNode' onClick={this.handleSubmit}>
            Create new
        </button>
      </div>
    );
  }
}

// Node requires "title" and "avatar",
// "-" button -> removes current NODE,
// "+" button -> calls Sidebar to create a new Node as a child
class Node extends Component {
  constructor(props) {
    super(props);
    this.state = {
      imageSrc: ''
    };
  }

  componentDidMount() {
    // Handle first rendering without ID from JSON data.
    const ID = !(this.props.ID)? 1 : this.props.ID;
    const avatarURL = "http://localhost:8080/getAvatar/ID=" + ID;

    fetch(avatarURL)
        .then(res => res.blob())
        .then(
          (result) => {
            const src = URL.createObjectURL(result);
            this.setState({
              imageSrc: src
            });
          },
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  handleAddClick(ID) {
    this.props.onAddClick(ID);
  }

  handleDeleteClick(ID) {
    this.props.onDeleteClick(ID);
  }

  render() {
    const ID = this.props.ID;

    // Return "-" button disabled for the ROOT NODE.
    const deleteButton = !(this.props.ID === 1) ?
    <button
      className="btn btn-deleteNode"
      disabled={this.props.buttonDisabled}
      onClick={() => this.handleDeleteClick(ID)} >
       {"-"}
    </button>
    :
    <button
      className="btn btn-deleteNode"
      disabled
      style={{color: "#282C34"}}  >
       {"-"}
    </button>

    return (
      <div className="Node-container"
        id={ID}
        parent-id={this.props.parentID} >
        <div className="Node">
          {deleteButton}
          <p>{this.props.title}</p>
          <button
            className="btn btn-addChild"
            disabled={this.props.buttonDisabled}
            onClick={() => this.handleAddClick(ID)} >
            {"+"}
          </button>
        </div>
        <img src={this.state.imageSrc} className="Node-Image"
            alt="Server error: no avatar uploaded" />
      </div>
    );
  }
}


// Tree - is tree structure,
// contains Node and its children, if any,
// children go as the same type (i.e. Tree),
// creating sub-Trees.
class Tree extends Component {
  constructor(props) {
    super(props);
    this.transferAddClick = this.transferAddClick.bind(this);
    this.transferDeleteClick = this.transferDeleteClick.bind(this);
  }
  // Transfer function calls from Node to TreeContainer.
  transferAddClick(ID) {
    this.props.onAddClick(ID);
  }
  transferDeleteClick(ID) {
    this.props.onDeleteClick(ID);
  }

  render() {
    // If Node has children -> recursion pattern
    let ChildNodes;
    if (this.props.node.ChildNodes != null) {
      ChildNodes = this.props.node.ChildNodes.map(
        (node, index)=>{
          return (<li key={node.ID} className="item">
                    <Tree node={node}
                          parentID={this.props.node.ID}
                          onAddClick={this.props.onAddClick}
                          onDeleteClick = {this.props.onDeleteClick}
                          buttonDisabled = {this.props.buttonDisabled}
                    />
                 </li>
          );
        }
      );

    }
    return(
      <div className="Tree" >
        <Node
          ID={this.props.node.ID}
          title={this.props.node.Title}
          parentID = {this.props.parentID}
          onAddClick = {this.transferAddClick}
          onDeleteClick = {this.transferDeleteClick}
          buttonDisabled = {this.props.buttonDisabled}
         />
         <ul className="container">
           {ChildNodes}
         </ul>
      </div>
    );
  }
}

// TreeContainer holds Tree and its sub-Trees,
// displays Sidebar, if Node's "+" button clicked,
// fetches JSON data from Backend,
// handles user actions from Nodes and Sidebar.
class TreeContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
      data: {}, //JSON data
      displaySidebar: false,
      buttonDisabled: false,
      currentNodeID: null
    }
    this.handleAddClick = this.handleAddClick.bind(this);
    this.handleDeleteClick = this.handleDeleteClick.bind(this);
    this.handleCreateClick = this.handleCreateClick.bind(this);
  }

  componentDidMount() {
    const url = 'http://localhost:8080/getTree';
    fetch(url)
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              data: result
            });
          },
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  handleAddClick(ID) {
    console.log(`Call to create a child with the parent node ID=${ID}`);
    this.setState(
      {displaySidebar: true,
       buttonDisabled: true,
       currentNodeID: ID
     });
  }

  handleCreateClick(file, title) {
    this.setState({displaySidebar: false, buttonDisabled:false});
    const ID = this.state.currentNodeID;
    const message = `Child appended to the node with ID=${ID}`;

    // Send data to Backend within formdata.
    let data = new FormData();
    // Check if avatar was uploaded and inform Backend about it.
    if (file != null) {
        data.append('filestatus', 'true');
    } else {
        data.append('filestatus', 'false');
    }

    data.append('uploadfile', file);
    const jsonData = JSON.stringify({
        ParentID: ID,
        Title: title
    });
    data.append('jsonData', jsonData);

    const url = 'http://localhost:8080/addNode';
    fetch(url, {
        method: 'POST',
        body: data
    })
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              data: result
            });
            console.log(message);
          },
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  handleDeleteClick(ID) {
    const url = "http://localhost:8080/deleteNode/ID=" + ID;
    fetch(url)
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              data: result
            });
            console.log(`Node with ID=${ID} - removed`)
          }
          ,
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  render() {
    // Display Sidebar depending on displaySidebar status.
    return(
      <div className="TreeContainer" >
        {this.state.displaySidebar &&
        <Sidebar
          onSubmit = {this.handleCreateClick}
        /> }

        <Tree
          node = {this.state.data}
          onAddClick = {this.handleAddClick}
          onDeleteClick = {this.handleDeleteClick}
          buttonDisabled = {this.state.buttonDisabled}
        />
      </div>
    );
  }
}

function Header(props) {
    return(
      <div className="App-header">
        <img src={props.logoSource} className="App-logo" alt="logo" />
        <h1 className="App-title">Welcome to React-my-tree project</h1>
      </div>
    );
}

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header logoSource={logo}  />
        <TreeContainer />
      </div>
    );
  }
}

export default App;
