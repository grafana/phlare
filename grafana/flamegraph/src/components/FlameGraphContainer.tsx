import React, { useState } from 'react';
import { DataFrame } from '@grafana/data';

import FlameGraphHeader from './FlameGraphHeader';
import FlameGraph from './FlameGraph/FlameGraph';
import { SelectedView } from './types';

type Props = {
  data: DataFrame;
};

const FlameGraphContainer = (props: Props) => {
  const [topLevelIndex, setTopLevelIndex] = useState(0);
  const [rangeMin, setRangeMin] = useState(0);
  const [rangeMax, setRangeMax] = useState(1);
  const [query, setQuery] = useState('');
  const [selectedView, setSelectedView] = useState(SelectedView.Both);

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
      />

      <FlameGraph
        data={props.data}
        topLevelIndex={topLevelIndex}
        rangeMin={rangeMin}
        rangeMax={rangeMax}
        query={query}
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        selectedView={selectedView}
        setSelectedView={setSelectedView}
      />
    </>
  );
};

export default FlameGraphContainer;
