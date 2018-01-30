import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import registerServiceWorker from './registerServiceWorker';

// You can use Console from browser DevTools to monitor user action messages
ReactDOM.render(<App />, document.getElementById('root'));
registerServiceWorker();
