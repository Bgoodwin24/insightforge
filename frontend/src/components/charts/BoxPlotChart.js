import React from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  Tooltip,
  Title,
} from 'chart.js';
import {
  BoxPlotController,
  BoxAndWiskers,
} from '@sgratzl/chartjs-chart-boxplot';
import { Chart } from 'react-chartjs-2';
import { generateColors, defaultOptions } from './utils/chartUtils';

ChartJS.register(
  BoxPlotController,
  BoxAndWiskers,
  CategoryScale,
  LinearScale,
  Tooltip,
  Title
);

export function BoxPlotChart({ labels, datasets, title }) {
  const colors = generateColors(datasets.length);

  const data = {
    labels,
    datasets: datasets.map((d, i) => ({
      label: d.label,
      data: d.data, // Should be an array of {min, q1, median, q3, max}
      backgroundColor: colors[i],
      borderColor: colors[i],
      borderWidth: 1,
    })),
  };

  const options = {
    ...defaultOptions(title),
    plugins: {
      tooltip: {
        callbacks: {
          label: (context) => {
            const { min, q1, median, q3, max } = context.raw;
            return `Min: ${min}, Q1: ${q1}, Median: ${median}, Q3: ${q3}, Max: ${max}`;
          },
        },
      },
    },
  };

  return <Chart type="boxplot" data={data} options={options} />;
}
