import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

// Sidebar element gets user input("title" and "avatar") for a new Node.
// Appears when "+" button is clicked on a Node.
class Sidebar extends Component {
  constructor(props) {
    super(props);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleSubmit() {
    const message = (this.fileInput.files[0]) ?
      `Selected file - ${this.fileInput.files[0].name}`
      :
      `No file uploaded`;
    alert(message);
    this.props.onSubmit();
  }

  handleTitleChange(e) {
    this.props.onTitleChange(e.target.value);
  }

  render() {
    return (
      <form className="Sidebar" onSubmit={this.handleSubmit}>
        <h3>Create Node </h3>
        <label>
          Upload file:
        </label>
        <input
          type="file"
          ref={input => {this.fileInput = input;}}
        />

        <input
          type="text"
          placeholder='Type a title'
          autoFocus
          onChange = {(e) => this.handleTitleChange(e)}
        />
        <button
          id='btn-createNode'
          onClick={this.handleSubmit}
           >
          Create
        </button>
      </form>
    );
  }
}

// Node element contains an avatar
// "-" button -> removes current NODE
// "+" button -> calls the Sidebar to append a new Node as a child
class Node extends Component {
  constructor(props) {
    super(props);
    this.state = {
      imageSrc: ''
    };
  }

  componentDidMount() {
    const url = "./images/react.png";
    fetch(url)
        .then(res => res.blob())
        .then(
          (result) => {
            const src = URL.createObjectURL(result);
            this.setState({
              imageSrc: src
            });
          }
          ,
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  handleAddClick(id) {
    this.props.onAddClick(id);
  }

  handleDeleteClick(parentId, id) {
    this.props.onDeleteClick(parentId, id);
  }

  render() {
    const id = this.props.id;
    const parentId = this.props.parentId;

    // Makes "-" button disabled for the ROOT NODE
    const deleteButton = !(this.props.id === 0) ?
    <button
      className="btn btn-deleteNode"
      disabled={this.props.buttonDisabled}
      onClick={() => this.handleDeleteClick(parentId, id)} >
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
        id={id}
        parent-id={this.props.parentId} >
        <div className="Node">
          {deleteButton}
          <p>{this.props.title}</p>
          <button
            className="btn btn-addChild"
            disabled={this.props.buttonDisabled}
            onClick={() => this.handleAddClick(id)} >
            {"+"}
          </button>
        </div>
        <img src={this.state.imageSrc} className="Node-Image" alt="Uploaded node avatar" />
      </div>
    );
  }
}

// Creates a tree depending on the JSON input data (treeData.json)
// If a Node element has children, Tree element works via recursion and creates a sub-tree
class Tree extends Component {
  constructor(props) {
    super(props);
    this.handleAddClick = this.handleAddClick.bind(this);
    this.handleDeleteClick = this.handleDeleteClick.bind(this);
  }

  handleAddClick(id) {
    this.props.onAddClick(id);
  }

  handleDeleteClick(parentId, id) {
    this.props.onDeleteClick(parentId, id);
  }

  render() {
    // If a Node element has children -> recursion pattern
    let childNodes;
    if (this.props.node.childNodes != null) {
      childNodes = this.props.node.childNodes.map(
        (node, index)=>{
          return (<li key={node.id} className="item">
                    <Tree node={node}
                          parentId={this.props.node.id}
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
          id={this.props.node.id}
          title={this.props.node.title}
          parentId = {this.props.parentId}
          onAddClick = {this.handleAddClick}
          onDeleteClick = {this.handleDeleteClick}
          buttonDisabled = {this.props.buttonDisabled}
         />
         <ul className="container">
           {childNodes}
         </ul>
      </div>
    );
  }
}

// TreeContainer element contains created Tree element and the Sidebar
// Fetches data from JSON -> treeData.json
// Gets and processes user actions from Node and Sidebar elements
class TreeContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
      data: {}, //JSON data
      displaySidebar: false,
      buttonDisabled: false,
      currentNodeId: null,
      currentNodeTitle: ''
    }
    this.handleAddClick = this.handleAddClick.bind(this);
    this.handleDeleteClick = this.handleDeleteClick.bind(this);
    this.handleCreateClick = this.handleCreateClick.bind(this);
    this.handleTitleChange = this.handleTitleChange.bind(this);
  }

  componentDidMount() {
    const url = "http://localhost:3001/api/treeData";
    fetch(url)
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              data: result
            });
          }
          ,
          (error) => {
            console.log(`Error: ${error.message}`);
          }
        )
  }

  handleAddClick(id) {
    console.log(`Call to create a child to the node with id=${id}`);
    this.setState(
      {displaySidebar: true,
       buttonDisabled: true,
       currentNodeId: id
     });
  }

  handleDeleteClick(parentId, id) {
    console.log(`Call to remove node with id=${id} from parentId=${parentId}`)
  }

  handleTitleChange(title) {
    this.setState({currentNodeTitle: title});
  }

  handleCreateClick() {
    this.setState({displaySidebar: false, buttonDisabled:false});
    const title = this.state.currentNodeTitle;
    const id = this.state.currentNodeId;
    const message = `Child with title: "${title}" appended to the node with id=${id}`;
    console.log(message);

  }

  render() {
    return(
      <div className="TreeContainer" >

        {this.state.displaySidebar &&
        <Sidebar
          onSubmit = {this.handleCreateClick}
          onTitleChange = {this.handleTitleChange}
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

// Page header
function Header(props) {
    return(
      <div className="App-header">
        <img src={props.logoSource} className="App-logo" alt="logo" />
        <h1 className="App-title">Welcome to React-my-tree project</h1>
      </div>
    );
}

// App element renders the whole app
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
