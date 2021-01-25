import React, { useEffect, useState, useRef } from 'react';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import _ from 'lodash';
import Select from 'react-select';
import { Link } from '@reach/router';
import { Button, Checkbox } from 'semantic-ui-react';
import { HuePicker } from 'react-color';

import styles from './index.module.scss';
import Service from '../../libs/service';
import Renderer from '../../libs/renderer';
import { COLOR } from '../../libs/constant';

const RenderSpacePage = (props) => {
  const mount = useRef(null);
  const control = useRef(null);

  const [renderingConfig, setRenderingConfig] = useState({
    group: null,
    scene: null,
    camera: null,
    renderer: null,
    rotating: true,
  });
  const [spaceId, setSpaceId] = useState();
  const [uiData, setUiData] = useState({
    spaceIdOptions: [],
    currentTime: 0,
    minNbOfEventsFilter: 0,
    colorMaskFilter: 3,
  });
  const [spaceTime, setSpaceTime] = useState();

  const { spaceId: defaultSpaceId } = props;

  const fetchSpaceIds = async () => {
    const spaces = await Service.getSpaces();

    const spaceIdOptions = spaces.map((space) => {
      const spaceId = space.space_id;
      return { value: spaceId, label: spaceId };
    });
    setUiData({ ...uiData, spaceIdOptions });
  };

  const getSpaceTime = async (spaceId) => {
    const { currentTime, minNbOfEventsFilter, colorMaskFilter } = uiData;
    const resp = await Service.getSpaceTime(
      spaceId,
      currentTime,
      minNbOfEventsFilter,
      colorMaskFilter,
    );

    debugger;
    const filteredNodes = _.get(resp, 'filtered_nodes', []);
    Renderer.clearGroup(renderingConfig.group);
    filteredNodes.forEach((node) => {
      Renderer.addPoint(renderingConfig.group, node.point);
    });
  };

  const onChangeSpaceId = async (option) => {
    const spaceId = option.value;
    setSpaceId(spaceId);
  };

  // componentDidMount, will load once only when page start
  useEffect(() => {
    fetchSpaceIds();

    const { clientWidth: width, clientHeight: height } = mount.current;

    const { group, scene, camera, renderer } = Renderer.init(width, height);
    setRenderingConfig({ ...renderingConfig, group, scene, camera, renderer });

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

  useEffect(() => {
    if (defaultSpaceId) {
      onChangeSpaceId({ value: spaceId });
    }
  }, [defaultSpaceId]);

  // called for every button clicks to update how the UI should render
  useEffect(() => {
    const { group, scene, camera, renderer, rotating } = renderingConfig;
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
  }, [renderingConfig]);

  return (
    <div className={styles.renderSpacePage}>
      <div className={styles.panel}>
        <div>
          <Button
            toggle
            active={renderingConfig.rotating}
            onClick={() =>
              setRenderingConfig({
                ...renderingConfig,
                rotating: !renderingConfig.rotating,
              })
            }
          >
            Rotate
          </Button>
        </div>
        <hr />
        <div>
          <span>Space ID: </span>
          <Select
            defaultValue={{
              value: defaultSpaceId,
              label: defaultSpaceId,
            }}
            onChange={onChangeSpaceId}
            options={uiData.spaceIdOptions}
            isSearchable={true}
          />
        </div>

        <div>
          <span>Current Time: </span>
          <input
            type="number"
            value={uiData.currentTime}
            onChange={(evt) => {
              setUiData({ ...uiData, currentTime: parseInt(evt.target.value) });
            }}
          />
        </div>
        <div>
          <span>Min # of events: </span>
          <input
            type="number"
            value={uiData.minNbOfEventsFilter}
            onChange={(evt) => {
              setUiData({
                ...uiData,
                minNbOfEventsFilter: parseInt(evt.target.value),
              });
            }}
          />
        </div>
        <div>
          <span>Color Mask Filter: </span>
          <input
            type="number"
            value={uiData.colorMaskFilter}
            onChange={(evt) => {
              setUiData({
                ...uiData,
                colorMaskFilter: parseInt(evt.target.value),
              });
            }}
          />
        </div>
        <hr />

        <Button
          icon="play"
          content="Render"
          labelPosition="left"
          disabled={!spaceId}
          onClick={() => getSpaceTime(spaceId)}
        />
      </div>
      <div className={styles.vis} ref={mount} />
    </div>
  );
};

export default RenderSpacePage;
