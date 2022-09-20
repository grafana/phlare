import React, { useCallback, useMemo, useState } from 'react';
import { DataFrame, DataFrameView } from '@grafana/data';
import { useWindowSize } from 'react-use';

import FlameGraphHeader from './FlameGraphHeader';
import FlameGraph from './FlameGraph/FlameGraph';
import { SelectedView } from './types';
import FlameGraphTopTable from './FlameGraph/FlameGraphTopTable';
import { MIN_WIDTH_TO_SHOW_TOP_TABLE } from '../constants';
import { Item, nestedSetToLevels } from './FlameGraph/dataTransform';

type Props = {
  data: DataFrame;
};

const FlameGraphContainer = (props: Props) => {
  const [topLevelIndex, setTopLevelIndex] = useState(0);
  const [rangeMin, setRangeMin] = useState(0);
  const [rangeMax, setRangeMax] = useState(1);
  const [query, setQuery] = useState('');
  const [selectedView, setSelectedView] = useState(SelectedView.Both);
  const { width: windowWidth } = useWindowSize();
  const profileTypeId = props.data.meta!.custom!.ProfileTypeID;

  // Transform dataFrame with nested set format to array of levels. Each level contains all the bars for a particular
  // level of the flame graph. We do this temporary as in the end we should be able to render directly by iterating
  // over the dataFrame rows.
  const levels = useMemo(() => {
    if (!props.data) {
      return [];
    }
    const dataView = new DataFrameView<Item>(props.data);
    return nestedSetToLevels(dataView);
  }, [props.data]);

  const renderTopTable = useCallback(() => {
    return (<FlameGraphTopTable levels={levels} profileTypeId={profileTypeId} selectedView={selectedView} />);
  }, [levels, profileTypeId, selectedView]);

  return (
    <>
      <FlameGraphHeader
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        query={query}
        setQuery={setQuery}
        selectedView={selectedView}
        setSelectedView={setSelectedView}
        windowWidth={windowWidth}
      />

      {selectedView !== SelectedView.FlameGraph && windowWidth >= MIN_WIDTH_TO_SHOW_TOP_TABLE ? renderTopTable() : null}

      <FlameGraph
        data={props.data}
        levels={levels}
        topLevelIndex={topLevelIndex}
        rangeMin={rangeMin}
        rangeMax={rangeMax}
        query={query}
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        selectedView={selectedView}
        setSelectedView={setSelectedView}
        windowWidth={windowWidth}
      />
    </>
  );
};

export default FlameGraphContainer;
