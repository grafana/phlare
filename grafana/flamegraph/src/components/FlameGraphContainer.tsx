import React, { useState } from 'react';

import FlameGraph from './FlameGraph';
import FlameGraphHeader from './FlameGraphHeader';
import { data } from '../data';

const FlameGraphContainer = () => {
  const flameGraphData = data['flamebearer'];
  const [topLevelIndex, setTopLevelIndex] = useState(0);
  const [rangeMin, setRangeMin] = useState(0);
  const [rangeMax, setRangeMax] = useState(1);
  const [query, setQuery] = useState('');

  return (
    <>
      <FlameGraphHeader
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        query={query}
        setQuery={setQuery}
      />

      <FlameGraph
        data={flameGraphData}
        topLevelIndex={topLevelIndex}
        rangeMin={rangeMin}
        rangeMax={rangeMax}
        query={query}
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
      />
    </>
  );
};

export default FlameGraphContainer;
