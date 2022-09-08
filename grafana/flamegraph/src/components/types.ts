export type TooltipData = {
  name: string,
  percentTitle: string,
  percentValue: number,
  unitTitle: string,
  unitValue: string,
  samples: number
}

export enum SampleUnit {
  Bytes = 'bytes',
  Count = 'count',
  Nanoseconds = 'nanoseconds'
}
