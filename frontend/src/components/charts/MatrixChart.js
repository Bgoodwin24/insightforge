import React from 'react';
import { Chart as ChartJS, CategoryScale, LinearScale, Tooltip, Title } from 'chart.js';
import { MatrixController, MatrixElement } from 'chartjs-chart-matrix';
import { Chart } from 'react-chartjs-2';
import { defaultOptions } from './utils/chartUtils';

ChartJS.register(MatrixController, MatrixElement, CategoryScale, LinearScale, Tooltip, Title);

export function heatmapColorScale(value) {
  if (typeof value !== "number") return "rgba(0, 0, 0, 0)"; // fallback for missing
  const r = Math.floor(255 * (1 - value));
  const b = Math.floor(255 * value);
  return `rgb(${r}, 0, ${b})`;
}

export function MatrixChart({ labels, datasets, title }) {
  const data = {
    labels, // Not used directly — labels come from the matrix structure
    datasets: datasets.map((d) => ({
      label: d.label,
      data: d.data, // Array of {x: colIndex, y: rowIndex, v: value}
      backgroundColor: (ctx) => heatmapColorScale(ctx.raw?.v),
      width: ({ chart }) =>
        chart.chartArea?.width ? chart.chartArea.width / labels.length : 10,
      height: ({ chart }) =>
        chart.chartArea?.height ? chart.chartArea.height / labels.length : 10,
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
          title: (ctx) => {
            const { x, y } = ctx[0]?.raw || {};
            return `(${x ?? "?"}, ${y ?? "?"})`;
          },
          label: (ctx) => {
            const v = ctx.raw?.v;
            return `Value: ${typeof v === "number" ? v.toFixed(3) : "N/A"}`;
          },
        },
      }
    },
  };

  return <Chart type='matrix' data={data} options={options} />;
}
