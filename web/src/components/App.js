import React, { useEffect, useState, useRef } from 'react';
import { Router, Link } from '@reach/router';

import PathContextList from './PathContextList';
import RenderPage from './RenderPage';

const App = () => {
  return (
    <Router>
      <PathContextList path="/" />
      <RenderPage path="render" />
    </Router>
  );
};

export default App;
