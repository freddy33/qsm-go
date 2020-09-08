import React, { useEffect, useState, useRef } from 'react';
import * as THREE from 'three';

const addAxes = (scene) => {
  const xAxis = new THREE.Line(
    new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(-50, 0, 0), new THREE.Vector3(50, 0, 0)]),
    new THREE.LineBasicMaterial({ color: 0xff0000 })
  );
  const yAxis = new THREE.Line(
    new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(0, -50, 0), new THREE.Vector3(0, 50, 0)]),
    new THREE.LineBasicMaterial({ color: 0x00ff00 })
  );
  const zAxis = new THREE.Line(
    new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(0, 0, -50), new THREE.Vector3(0, 0, 50)]),
    new THREE.LineBasicMaterial({ color: 0x0000ff })
  );

  scene.add(xAxis);
  scene.add(yAxis);
  scene.add(zAxis);
};

const addPoint = (scene) => {
  const geometry = new THREE.SphereGeometry(1, 32, 32);
  const material = new THREE.MeshBasicMaterial({ color: 0xffff00 });
  const sphere = new THREE.Mesh(geometry, material);
  scene.add(sphere);
};

const App = () => {
  const mount = useRef(null);

  useEffect(() => {
    let width = mount.current.clientWidth;
    let height = mount.current.clientHeight;
    let frameId;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, width / height, 1, 500);
    // const camera = new THREE.PerspectiveCamera( 45, window.innerWidth / window.innerHeight, 1, 500 );
    camera.position.set(45, 90, 100);
    camera.lookAt(0, 0, 0);
    const renderer = new THREE.WebGLRenderer({ antialias: true });

    // const geometry = new THREE.BoxGeometry();
    // const material = new THREE.MeshBasicMaterial({ color: 0x00ff00 });
    // const cube = new THREE.Mesh(geometry, material);
    // scene.add(cube);
    // camera.position.z = 5;

    addAxes(scene);
    addPoint(scene);

    // renderer.setClearColor('#000000');
    renderer.setSize(width, height);

    // const renderScene = () => {
    //   renderer.render(scene, camera);
    // };

    // const handleResize = () => {
    //   width = mount.current.clientWidth;
    //   height = mount.current.clientHeight;
    //   renderer.setSize(width, height);
    //   camera.aspect = width / height;
    //   camera.updateProjectionMatrix();
    //   renderScene();
    // };

    // const animate = () => {
    //   // cube.rotation.x += 0.01;
    //   // cube.rotation.y += 0.01;

    //   renderScene();
    //   frameId = window.requestAnimationFrame(animate);
    // };

    const animate = () => {
      // cube.rotation.x += 0.01;
      // cube.rotation.y += 0.01;
      // line.rotation.x += 0.01;
      // line.rotation.y += 0.01;

      requestAnimationFrame(animate);
      renderer.render(scene, camera);
    };

    animate();

    mount.current.appendChild(renderer.domElement);
    // window.addEventListener('resize', handleResize);

    // requestAnimationFrame(animate);
  }, []);

  return <div className="vis" ref={mount} />;
};

export default App;
