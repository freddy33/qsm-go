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

export default {
  addLine,
  addAxes,
  addPoint,
  connectPoints,
  addCameraPivot,
};
