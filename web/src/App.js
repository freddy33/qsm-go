import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';
import _ from 'lodash';
import Popup from 'reactjs-popup';
import 'reactjs-popup/dist/index.css';

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
    Service.fetchPointPackDataMsg().then((pointPackDataMsg) => {
      setPointPackDataMsg(convertPointPackDataMsgToState(pointPackDataMsg));
    });

    let width = mount.current.clientWidth;
    let height = mount.current.clientHeight;

    const { scene, camera, renderer } = Renderer.init(width, height);
    setScene(scene);
    setCamera(camera);
    setRenderer(renderer);

    mount.current.appendChild(renderer.domElement);

    const handleResize = () => {
      width = mount.current.clientWidth;
      height = mount.current.clientHeight;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.render(scene, camera);
    };

    window.addEventListener('resize', handleResize);
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
    <div>
      <div className="vis" ref={mount} />

      <Popup trigger={<button className="control"> Configure </button>} modal>
        {(close) => (
          <div className="configure">
            {/* <button onClick={() => setRotating(!rotating)}>Rotate</button> */}
            <div>
              <textarea
                onChange={(evt) => {
                  setDataInput(_.get(evt, 'target.value', ''));
                }}
                rows="40"
                value={JSON.stringify(Renderer.mockPoints)}
              ></textarea>
            </div>
            <div>
              <button
                onClick={() => {
                  const data = JSON.parse(dataInput);
                  Renderer.draw(scene, data, pointPackDataMsg);
                  close();
                }}
              >
                Load data
              </button>
            </div>
          </div>
        )}
      </Popup>
    </div>
  );
};

export default App;
