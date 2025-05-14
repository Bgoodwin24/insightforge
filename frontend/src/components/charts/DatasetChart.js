import React from 'react';
import { BarChart } from './BarChart';
import { LineChart } from './LineChart';
import { StackedBarChart } from './StackedBarChart';
import { BoxPlotChart } from './BoxPlotChart';
import { MatrixChart } from './MatrixChart';

export default function DatasetChart({ chartType, labels, datasets, title }) {
  switch (chartType) {
    case 'bar':
      return <BarChart labels={labels} datasets={datasets} title={title} />;
    case 'line':
      return <LineChart labels={labels} datasets={datasets} title={title} />;
    case 'stackedBar':
      return <StackedBarChart labels={labels} datasets={datasets} title={title} />;
    case 'boxplot':
      return <BoxPlotChart labels={labels} datasets={datasets} title={title} />;
    case 'matrix':
      return <MatrixChart labels={labels} datasets={datasets} title={title} />;
      default:
      return <div>Unsupported chart type: {chartType}</div>;
  }
}
