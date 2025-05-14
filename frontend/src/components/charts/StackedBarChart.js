import React from 'react';
import { Bar } from 'react-chartjs-2';
import { generateColors, defaultOptions } from './utils/chartUtils';

export function StackedBarChart({ labels, datasets, title }) {
  const colors = generateColors(datasets.length);

  const data = {
    labels,
    datasets: datasets.map((d, i) => ({
      label: d.label,
      data: d.data,
      backgroundColor: colors[i],
      stack: 'Stack 0',
    })),
  };

  const options = {
    ...defaultOptions(title),
    scales: {
      x: { stacked: true },
      y: { stacked: true },
    },
  };

  return <Bar data={data} options={options} />;
}
