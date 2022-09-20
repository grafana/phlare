import { css } from '@emotion/css';
import React from 'react';

import { Button, Input, useStyles, RadioButtonGroup } from '@grafana/ui';

import { SelectedView } from './types';
import { MIN_WIDTH_TO_SHOW_TOP_TABLE } from '../constants';

type Props = {
  query: string;
  setTopLevelIndex: (level: number) => void;
  setRangeMin: (range: number) => void;
  setRangeMax: (range: number) => void;
  setQuery: (query: string) => void;
  selectedView: SelectedView;
  setSelectedView: (view: SelectedView) => void;
  windowWidth: number;
};

const viewOptions: Array<{ value: string; label: string; description: string }> = [
  { value: SelectedView.TopTable, label: 'Top Table', description: 'Only show top table' },
  { value: SelectedView.FlameGraph, label: 'Flame Graph', description: 'Only show flame graph' },
  { value: SelectedView.Both, label: 'Both', description: 'Show both the top table and flame graph' },
];

const FlameGraphHeader = ({
  query,
  setTopLevelIndex,
  setRangeMin,
  setRangeMax,
  setQuery,
  selectedView,
  setSelectedView,
  windowWidth,
}: Props) => {
  const styles = useStyles(getStyles);

  return (
    <div className={styles.header}>
      <div className={styles.leftContainer}>
        <div className={styles.inputContainer}>
          <Input
            value={query || ''}
            onChange={(v) => {
              setQuery(v.currentTarget.value);
            }}
            placeholder={'Search..'}
            width={24}
          />
        </div>
        <Button
          type={'button'}
          variant={'secondary'}
          size={'md'}
          onClick={() => {
            setTopLevelIndex(0);
            setRangeMin(0);
            setRangeMax(1);
            setQuery('');
          }}
        >
          Reset View
        </Button>
      </div>

      {windowWidth >= MIN_WIDTH_TO_SHOW_TOP_TABLE && (
        <div className={styles.rightContainer}>
          <RadioButtonGroup
            options={viewOptions}
            value={selectedView}
            onChange={(view) => {
              setSelectedView(view as SelectedView);
            }}
          />
        </div>
      )}
    </div>
  );
};

const getStyles = () => ({
  header: css`
    display: flow-root;
    padding: 20px 0;
    width: 100%;
  `,
  inputContainer: css`
    float: left;
    margin-right: 20px;
  `,
  leftContainer: css`
    float: left;
  `,
  rightContainer: css`
    float: right;
  `,
});

export default FlameGraphHeader;
