import React, { Component } from 'react';
// import ReactDOM from 'react-dom';
import logo from './logo.svg';
import './App.css';

class Sidebar extends Component {
  constructor(props) {
    super(props);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleSubmit() {
    this.props.onSubmit();
  }

  handleTitleChange(e) {
    this.props.onTitleChange(e.target.value);
  }

  render() {
    return (
      <form className="Sidebar" onSubmit={this.handleSubmit}>
        <h3>Create Node </h3>
          <input
            type="text"
            placeholder='Type a title'
            autoFocus
            onChange = {(e) => this.handleTitleChange(e)}
          />
        <input
          id='btn-createNode'
          type="submit"
          value="Create"
        />
      </form>
    );
  }
}

class Node extends Component {
  constructor(props) {
    super(props);
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
    const deleteButton = !(this.props.id === 0) ?
    <button
      className="btn btn-deleteNode"
      disabled={this.props.buttonDisabled}
      onClick={() => this.handleDeleteClick(parentId, id)} >
       {"-"}
    </button>
    :
    <button style={{width: '2rem'}} className="btn btn-deleteNode" />;

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
        <img src={require('./images/react.png')} className="Node-Image" alt="node-image" />
      </div>
    );
  }
}

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

class TreeContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
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
          node = {treeData}
          onAddClick = {this.handleAddClick}
          onDeleteClick = {this.handleDeleteClick}
          buttonDisabled = {this.state.buttonDisabled}
        />
      </div>
    );
  }
}

const treeData =
  {
    id: 0,
    title: "ROOT",
    childNodes: [
      {
        id: 1,
        title: 'NODE 1',
        childNodes: [
          {
            id: 3,
            title: "NODE 3",
            childNodes: []
          },
          {
            id: 4,
            title: "NODE 4",
            childNodes: []
          }
        ]
      },
      {
        id: 2,
        title: "NODE 2",
        childNodes: [
          {
            id: 5,
            title: "NODE 5",
            childNodes: [
              {
                id: 7,
                title: "NODE 7",
                childNodes: []
              }
            ]
          }
        ]
      },
      {
        id: 6,
        title: "NODE 6",
        childNodes: []
      }
    ]
  }
;

function Header(props) {
    return(
      <div className="App-header">
        <img src={props.logoSource} className="App-logo" alt="logo" />
        <h1 className="App-title">Welcome to my-tree project</h1>
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
