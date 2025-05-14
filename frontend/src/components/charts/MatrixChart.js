import React from 'react';
import { Chart as ChartJS, CategoryScale, LinearScale, Tooltip, Title } from 'chart.js';
import { MatrixController, MatrixElement } from 'chartjs-chart-matrix';
import { Chart } from 'react-chartjs-2';
import { defaultOptions } from './utils/chartUtils';

ChartJS.register(MatrixController, MatrixElement, CategoryScale, LinearScale, Tooltip, Title);

export function MatrixChart({ labels, datasets, title }) {
  const data = {
    labels, // Not used directly — labels come from the matrix structure
    datasets: datasets.map((d) => ({
      label: d.label,
      data: d.data, // Array of {x: colIndex, y: rowIndex, v: value}
      backgroundColor: (ctx) => {
        const value = ctx.raw.v;
        // Simple red-blue scale
        const r = Math.floor(255 * (1 - value));
        const b = Math.floor(255 * value);
        return `rgb(${r}, 0, ${b})`;
      },
      width: ({ chart }) => chart.chartArea.width / labels.length,
      height: ({ chart }) => chart.chartArea.height / labels.length,
    })),
  };

  const options = {
    ...defaultOptions(title),
    scales: {
      x: {
        type: 'category',
        labels,
        offset: true,
        grid: { display: false },
      },
      y: {
        type: 'category',
        labels: labels.slice().reverse(), // so it reads top-down
        offset: true,
        grid: { display: false },
      },
    },
    plugins: {
      tooltip: {
        callbacks: {
          title: (ctx) => `(${ctx[0].raw.x}, ${ctx[0].raw.y})`,
          label: (ctx) => `Value: ${ctx.raw.v}`,
        },
      },
    },
  };

  return <Chart type='matrix' data={data} options={options} />;
}
