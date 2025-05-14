import React from 'react';
import { Line } from 'react-chartjs-2';
import { generateColors, defaultOptions } from './utils/chartUtils';

export function LineChart({ labels, datasets, title }) {
  const colors = generateColors(datasets.length);

  const data = {
    labels,
    datasets: datasets.map((d, i) => ({
      label: d.label,
      data: d.data,
      borderColor: colors[i],
      backgroundColor: colors[i] + '33',
      fill: false,
    })),
  };

  return <Line data={data} options={defaultOptions(title)} />;
}
