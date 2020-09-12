import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';

const addLine = (scene, from, to, color = 0xffff00) => {
  const line = new THREE.Line(
    new THREE.BufferGeometry().setFromPoints([
      new THREE.Vector3(from.x, from.y, from.z),
      new THREE.Vector3(to.x, to.y, to.z),
    ]),
    new THREE.LineBasicMaterial({ color })
  );

  scene.add(line);
};

const addAxes = (scene) => {
  addLine(scene, { x: -50, y: 0, z: 0 }, { x: 50, y: 0, z: 0 }, 0xff0000);
  addLine(scene, { x: 0, y: -50, z: 0 }, { x: 0, y: 50, z: 0 }, 0x00ff00);
  addLine(scene, { x: 0, y: 0, z: -50 }, { x: 0, y: 0, z: 50 }, 0x0000ff);
};

const addPoint = (scene, pos) => {
  const geometry = new THREE.SphereGeometry(1, 32, 32);
  const material = new THREE.MeshBasicMaterial({ color: 0xffff00 });
  const sphere = new THREE.Mesh(geometry, material);
  sphere.position.set(pos.x, pos.y, pos.z);
  scene.add(sphere);
};

const connectPoints = (scene, from, to, color) => {
  addLine(scene, from, to);
  addPoint(scene, to);
};

const addCameraPivot = (scene, camera) => {
  const cameraPivot = new THREE.Object3D();

  scene.add(cameraPivot);
  cameraPivot.add(camera);
  camera.position.set(65, 90, 100);
  camera.lookAt(cameraPivot.position);

  return cameraPivot;
};

const App = () => {
  const mount = useRef(null);

  const [rotating, setRotating] = useState(true);
  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();

  // let scene, camera, renderer;

  useEffect(() => {
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

    addAxes(scene);

    const origin = { x: 0, y: 0, z: 0 };
    addPoint(scene, origin);
    connectPoints(scene, origin, { x: 10, y: 10, z: 10 }, 0xffff00);
    connectPoints(scene, origin, { x: -10, y: 10, z: 10 }, 0xffff00);
    connectPoints(scene, origin, { x: 10, y: -10, z: 10 }, 0xffff00);

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

  useEffect(() => {
    if (!(scene && camera && renderer)) return;

    const cameraPivot = addCameraPivot(scene, camera);
    const animate = () => {
      if (rotating) {
        cameraPivot.rotateOnAxis(new THREE.Vector3(0, 1, 0), 0.01);
      }

      requestAnimationFrame(animate);
      renderer.render(scene, camera);
    };

    animate();
  }, [rotating, scene, camera, renderer]);

  const toggleRotation = () => {
    console.log('fuck');
    setRotating(!rotating);
  };

  return (
    <div>
      <div className="vis" ref={mount} />
      <button className="control" onClick={() => toggleRotation()}>
        Rotate
      </button>
      
    </div>
  );
};

export default App;
