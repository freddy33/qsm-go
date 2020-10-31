import React, { useEffect, useState, useRef } from 'react';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import _ from 'lodash';
import Select from 'react-select';

import Service from './service';
import Renderer from './renderer';

const App = () => {
  const mount = useRef(null);
  const control = useRef(null);

  const [rotating, setRotating] = useState(true);
  const [group, setGroup] = useState();
  const [scene, setScene] = useState();
  const [camera, setCamera] = useState();
  const [renderer, setRenderer] = useState();
  const [pathContextIdOptions, setPathContextIdOptions] = useState([]);
  const [currentPathContext, setCurrentPathContext] = useState({});
  const [fromDist, setFromDist] = useState(0);
  const [toDist, setToDist] = useState(0);

  const fetchPathContextIds = async () => {
    const pathContextIds = await Service.getPathContextIds();

    const pathContextIdOptions = pathContextIds.map((pathContextId) => {
      return { value: pathContextId, label: pathContextId };
    });
    setPathContextIdOptions(pathContextIdOptions);
  };

  const updateMaxDist = async () => {
    const { pathContextId, maxDist } = currentPathContext;
    await Service.updateMaxDist(pathContextId, maxDist + 1);

    const pathContext = await Service.getPathContext(pathContextId);
    setCurrentPathContext(pathContext);
  };

  const getPathNodes = async () => {
    if (fromDist > toDist) {
      alert('"From" dist cannot be less than "To" dist');
      return;
    }

    if (toDist > currentPathContext.maxDist) {
      alert(`"To" dist needs to be less than ${currentPathContext.maxDist}`);
      return;
    }

    const resp = await Service.getPathNodes(currentPathContext.pathContextId, fromDist, toDist);
    const pathNodes = _.get(resp, 'path_nodes', []);

    if (!pathNodes) {
      alert(resp);
    }

    const sortByDist = _.sortBy(pathNodes, ['d']);

    const nodeMap = {};
    sortByDist.forEach((node) => {
      const linkedPathNodeIds = _.get(node, 'linked_path_node_ids', []);

      linkedPathNodeIds.forEach((precedingNodeId) => {
        const precedingNode = nodeMap[precedingNodeId];

        if (!precedingNode) {
          return;
        }

        const childNodes = _.get(precedingNode, 'childNodes', []);
        childNodes.push(node);
        precedingNode.childNodes = childNodes;
      });

      const pathNodeId = node.path_node_id;
      nodeMap[pathNodeId] = node;
    });

    const nodesToDraw = _.filter(nodeMap, { d: fromDist });

    Renderer.drawRoots(group, nodesToDraw);
  };

  const onChangePathContextId = async (option) => {
    const pathContextId = option.value;
    const pathContext = await Service.getPathContext(pathContextId);

    setCurrentPathContext(pathContext);
  };

  // componentDidMount, will load once only when page start
  useEffect(() => {
    fetchPathContextIds();

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { group, scene, camera, renderer } = Renderer.init(width, height);
    setGroup(group);
    setScene(scene);
    setCamera(camera);
    setRenderer(renderer);

    const handleResize = () => {
      const { clientWidth: width, clientHeight: height } = mount.current;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
      renderer.render(scene, camera);
    };

    new OrbitControls(camera, renderer.domElement);
    mount.current.appendChild(renderer.domElement);
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    if (!(scene && camera && renderer && group)) return;

    const frameId = _.get(control, 'current.frameId');
    if (frameId) {
      cancelAnimationFrame(frameId);
    }

    const animate = () => {
      if (rotating) {
        group.rotation.y += 0.005;
      }
      const frameId = requestAnimationFrame(animate);
      control.current = { frameId };
      renderer.render(scene, camera);
    };

    animate();
  }, [scene, camera, renderer, group, rotating]);

  return (
    <div className="main">
      <div className="panel">
        <div>
          <button onClick={() => setRotating(!rotating)}>Rotate: {`${rotating}`}</button>
        </div>
        <hr />
        <div>
          <span>Path Context ID:</span>
          <Select onChange={onChangePathContextId} options={pathContextIdOptions} isSearchable={true} />
        </div>

        <div>
          <p>Growth Type: {currentPathContext.growthType} </p>
          <p>Growth Index: {currentPathContext.growthIndex} </p>
          <p>Growth Offset: {currentPathContext.growthOffset} </p>
          <p>Max Dist: {currentPathContext.maxDist}</p>
        </div>
        <hr />
        <div>
          <button disabled={!currentPathContext.pathContextId} onClick={() => updateMaxDist()}>
            Max Dist + 1
          </button>
        </div>
        <hr />
        <div>
          <div>
            <span>From Dist: </span>
            <input
              type="number"
              value={fromDist}
              onChange={(evt) => {
                setFromDist(parseInt(evt.target.value));
              }}
            />
          </div>
          <div>
            <span>To Dist: </span>
            <input
              type="number"
              value={toDist}
              onChange={(evt) => {
                setToDist(parseInt(evt.target.value));
              }}
            />
          </div>
          <div>
            <button disabled={!currentPathContext.pathContextId} onClick={() => getPathNodes()}>
              Render
            </button>
          </div>

          <hr />
        </div>
      </div>
      <div className="vis" ref={mount} />
    </div>
  );
};

export default App;
