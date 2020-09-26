import * as THREE from 'three';

const init = (width, height) => {
  const camera = new THREE.PerspectiveCamera(45, width / height, 1, 500);
  camera.position.set(45, 90, 100);
  camera.lookAt(0, 0, 0);

  const renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setSize(width, height);

  const scene = new THREE.Scene();
  addAxes(scene);

  return {
    scene,
    camera,
    renderer,
  };
};

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
  addLine(scene, { x: -70, y: 0, z: 0 }, { x: 70, y: 0, z: 0 }, 0xff0000);
  addLine(scene, { x: 0, y: -70, z: 0 }, { x: 0, y: 70, z: 0 }, 0x00ff00);
  addLine(scene, { x: 0, y: 0, z: -70 }, { x: 0, y: 0, z: 70 }, 0x0000ff);
};

const addPoint = (scene, pos) => {
  const geometry = new THREE.SphereGeometry(0.5, 32, 32);
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

const getRandomInt = (max) => Math.floor(Math.random() * Math.floor(max));
const mockPoints = [...Array(10).keys()].map((i) => {
  return {
    id: i,
    x: i + getRandomInt(10),
    y: i + getRandomInt(10),
    z: i + getRandomInt(10),
    trioId: i + getRandomInt(10),
  };
});

const draw = (scene, points, pointPackDataMsg) => {
  if (!pointPackDataMsg) {
    return;
  }

  const { connections, trios } = pointPackDataMsg;

  points.forEach((point) => {
    const startingPoint = { x: point.x, y: point.y, z: point.z };
    addPoint(scene, startingPoint);
    const trio = trios[point.trioId];
    trio.connIds.forEach((connId) => {
      const trio = connections[connId];
      const connPoint = {
        x: startingPoint.x + trio.vector.x,
        y: startingPoint.y + trio.vector.y,
        z: startingPoint.z + trio.vector.z,
      };
      addPoint(scene, connPoint);
      addLine(scene, startingPoint, connPoint);
    });
  });
};

export default {
  init,
  addLine,
  addAxes,
  addPoint,
  connectPoints,
  addCameraPivot,
  draw,
  mockPoints,
};
