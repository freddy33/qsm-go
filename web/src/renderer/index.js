import * as THREE from 'three';
import _ from 'lodash';

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

const drawRoots = (scene, roots) => {
  console.time('drawRoots');
  if (!_.isArray(roots)) {
    return;
  }

  roots.forEach((root) => {
    drawRoot(scene, root);
  });

  console.timeEnd('drawRoots');
};

const drawRoot = (scene, originalRoot) => {
  const root = _.cloneDeep(originalRoot);
  const stack = [];
  let current = root;

  stack.push(current);
  do {
    while (current) {
      const childNodes = _.get(current, 'childNodes', []);
      if (!childNodes.length) {
        break;
      }

      current = childNodes.pop();
      stack.push(current);
    }

    if (stack.length > 0) {
      const node = stack.pop();

      const point = _.get(node, 'point');
      addPoint(scene, point);
      const parent = _.last(stack);
      if (parent) {
        const parentPoint = _.get(parent, 'point');
        addLine(scene, parentPoint, point);
      }

      current = parent;
    }
  } while (current || stack.length > 0);
};

export default {
  init,
  addLine,
  addAxes,
  addPoint,
  connectPoints,
  addCameraPivot,
  drawRoot,
  drawRoots,
};
