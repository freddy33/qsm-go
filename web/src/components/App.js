import React from 'react';
import { Router } from '@reach/router';

import PathContextList from './PathContextList';
import RenderPage from './RenderPage';

const App = () => {
  return (
    <Router>
      <PathContextList path="/" />
      <RenderPage path="render" />
      <RenderPage path="render/:pathContextId" />
    </Router>
  );
};

export default App;
