import React from 'react';
import { Bar } from 'react-chartjs-2';
import { generateColors, defaultOptions } from './utils/chartUtils';

export function BarChart({ labels, datasets, title }) {
  const colors = generateColors(datasets.length);

  const data = {
    labels,
    datasets: datasets.map((d, i) => ({
      label: d.label,
      data: d.data,
      backgroundColor: colors[i],
      barThickness: 100, // Force visible width even on single label
      maxBarThickness: 120, // Prevent overflow
    })),
  };

  return <Bar data={data} options={defaultOptions(title)} />;
}
