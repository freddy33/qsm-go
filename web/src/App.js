import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';

import Service from './service';
import Renderer from './renderer';

const App = () => {
  const mount = useRef(null);

  const [rotating, setRotating] = useState(true);
  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();
  const [pointPackDataMsg, setPointPackDataMsg] = useState();

  // componentDidMount, will load once only when page start
  useEffect(() => {
    Service.fetchPointPackDataMsg().then((pointPackDataMsg) => {
      debugger;
      setPointPackDataMsg(pointPackDataMsg);
    });

    let width = mount.current.clientWidth;
    let height = mount.current.clientHeight;

    const scene = new THREE.Scene();
    setScene(scene);
    const camera = new THREE.PerspectiveCamera(45, width / height, 1, 500);
    setCamera(camera);

    camera.position.set(45, 90, 100);
    camera.lookAt(0, 0, 0);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    setRenderer(renderer);

    Renderer.addAxes(scene);

    const origin = { x: 0, y: 0, z: 0 };
    Renderer.addPoint(scene, origin);
    Renderer.connectPoints(scene, origin, { x: 10, y: 10, z: 10 }, 0xffff00);
    Renderer.connectPoints(scene, origin, { x: -10, y: 10, z: 10 }, 0xffff00);
    Renderer.connectPoints(scene, origin, { x: 10, y: -10, z: 10 }, 0xffff00);

    renderer.setSize(width, height);

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
      <button className="control" onClick={() => setRotating(!rotating)}>
        Rotate
      </button>
    </div>
  );
};

export default App;
