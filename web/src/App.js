import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';
import _ from 'lodash';

import Service from './service';
import Renderer from './renderer';

const convertPointPackDataMsgToState = (pointPackDataMsg) => {
  const connections = {};
  const trios = {};
  pointPackDataMsg.getAllConnectionsList().forEach((conn) => {
    connections[conn.getConnId()] = {
      connId: conn.getConnId(),
      ds: conn.getDs(),
      vector: {
        x: conn.getVector().getX(),
        y: conn.getVector().getY(),
        z: conn.getVector().getZ(),
      },
    };
  });

  pointPackDataMsg.getAllTriosList().forEach((trio) => {
    trios[trio.getTrioId()] = {
      trioId: trio.getTrioId(),
      connIds: trio.getConnIdsList(),
    };
  });

  return { connections, trios };
};

const App = () => {
  const mount = useRef(null);

  const [rotating, setRotating] = useState(true);
  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();
  const [pointPackDataMsg, setPointPackDataMsg] = useState();
  const [dataInput, setDataInput] = useState('');

  // componentDidMount, will load once only when page start
  useEffect(() => {
    Service.fetchPointPackDataMsg().then((response) => {
      const pointPackDataMsg = convertPointPackDataMsgToState(response);
      setPointPackDataMsg(pointPackDataMsg);
    });

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { scene, camera, renderer } = Renderer.init(width, height);
    setScene(scene);
    setCamera(camera);
    setRenderer(renderer);

    mount.current.appendChild(renderer.domElement);

    const handleResize = () => {
      const { clientWidth: width, clientHeight: height } = mount.current;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.render(scene, camera);
    };

    window.addEventListener('resize', handleResize);

    setDataInput(JSON.stringify(Renderer.mockPoints));
  }, []);

  useEffect(() => {
    Renderer.draw(scene, Renderer.mockPoints, pointPackDataMsg);
  }, [pointPackDataMsg]);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    if (!(scene && camera && renderer)) return;

    const cameraPivot = Renderer.addCameraPivot(scene, camera);
    const animate = () => {
      if (rotating) {
        cameraPivot.rotateOnAxis(new THREE.Vector3(0, 1, 0), 0.01);
      }

      requestAnimationFrame(animate);
      renderer.render(scene, camera);
    };

    animate();
  }, [rotating, scene, camera, renderer]);

  return (
    <div className="main">
      <div className="panel">
        <div>
          <textarea
            onChange={(evt) => {
              setDataInput(_.get(evt, 'target.value', ''));
            }}
            rows="30"
            value={dataInput}
          />
        </div>
        <div>
          <button
            onClick={() => {
              const data = JSON.parse(dataInput);

              const { clientWidth: width, clientHeight: height } = mount.current;

              const { scene, camera, renderer: newRenderer } = Renderer.init(width, height);
              setScene(scene);
              setCamera(camera);
              setRenderer(newRenderer);

              mount.current.replaceChild(newRenderer.domElement, renderer.domElement);
              Renderer.draw(scene, data, pointPackDataMsg);
            }}
          >
            Load data
          </button>
        </div>
        <div>
          <button onClick={() => setRotating(!rotating)}>Rotate</button>
        </div>
      </div>
      <div className="vis" ref={mount} />
    </div>
  );
};

export default App;
