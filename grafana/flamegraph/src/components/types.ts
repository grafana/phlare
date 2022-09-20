export type TooltipData = {
  name: string,
  percentTitle: string,
  percentValue: number,
  unitTitle: string,
  unitValue: string,
  samples: string
}

export enum SampleUnit {
  Bytes = 'bytes',
  Count = 'count',
  Nanoseconds = 'nanoseconds'
}

export enum SelectedView {
  TopTable = 'topTable',
  FlameGraph = 'flameGraph',
  Both = 'both'
}

export const BYTE_UNITS = [
  { divider: 1024, suffix: 'KB' },
  { divider: 1024, suffix: 'MB' },
  { divider: 1024, suffix: 'GB' },
  { divider: 1024, suffix: 'PT' },
];

export const COUNT_UNITS = [
  { divider: 1000, suffix: 'K' },
  { divider: 1000, suffix: 'M' },
  { divider: 1000, suffix: 'G' },
  { divider: 1000, suffix: 'T' },
];

export const NANOSECOND_UNITS = [
  { divider: 60, suffix: 'minutes' },
  { divider: 60, suffix: 'hours' },
  { divider: 24, suffix: 'days' },
];
