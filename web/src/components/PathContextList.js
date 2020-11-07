import React, { useEffect, useState, useRef } from 'react';
import { Link } from '@reach/router';

const PathContextList = () => {
  return (
    <div>
      <h1>Home</h1>
      <nav>
        <Link to="/render">Render</Link> | <Link to="dashboard">Dashboard</Link>
      </nav>
    </div>
  );
};

export default PathContextList;
