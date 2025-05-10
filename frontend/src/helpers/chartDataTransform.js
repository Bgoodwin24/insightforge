/**
 * Transforms a map of group-value pairs into Chart.js format.
 *
 * @param {Object} data - The backend response (e.g., { "GroupA": 123, "GroupB": 456 }).
 * @param {String} datasetLabel - The label to display on the chart legend.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformGroupedDataToChartJS(data, datasetLabel = "Result") {
    const labels = Object.keys(data);
    const values = Object.values(data);
  
    return {
      labels,
      datasets: [
        {
          label: datasetLabel,
          data: values,
          backgroundColor: "rgba(54, 162, 235, 0.5)",
          borderColor: "rgba(54, 162, 235, 1)",
          borderWidth: 1,
        },
      ],
    };
  }

/**
 * Transforms a pivot table response into Chart.js format.
 *
 * @param {Object} pivotTable - A nested object like:
 *   {
 *     "RowA": { "Col1": 10, "Col2": 20 },
 *     "RowB": { "Col1": 15, "Col2": 25 }
 *   }
 * @returns {Object} - Chart.js config with `labels` and `datasets`
 */
function transformPivotDataToChartJS(pivotTable) {
    const rowLabels = Object.keys(pivotTable);
    const columnSet = new Set();
  
    // Collect all unique column keys
    rowLabels.forEach(row => {
      const columns = pivotTable[row];
      Object.keys(columns).forEach(col => columnSet.add(col));
    });
  
    const columnLabels = Array.from(columnSet).sort(); // sorted for consistency
  
    const datasets = columnLabels.map(col => {
      const data = rowLabels.map(row => {
        const value = pivotTable[row][col];
        return value !== undefined ? value : null;
      });
  
      return {
        label: col,
        data,
        backgroundColor: randomRGBA(), // Optional: assign a unique color per column
        borderColor: "rgba(0,0,0,0.1)",
        borderWidth: 1
      };
    });
  
    return {
      labels: rowLabels,
      datasets
    };
  }
  
  /**
   * Optional: Generate a random RGBA color for dataset styling.
   */
function randomRGBA() {
    const r = Math.floor(Math.random() * 156) + 100;
    const g = Math.floor(Math.random() * 156) + 100;
    const b = Math.floor(Math.random() * 156) + 100;
    return `rgba(${r}, ${g}, ${b}, 0.6)`;
  }

/**
 * Transforms the cleaned rows after missing values are dropped into Chart.js format.
 *
 * @param {Object} result - The backend response with `rows` (cleaned rows).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformDroppedRowsToChartJS(result, datasetLabel = "Cleaned Dataset") {
    return {
        labels: result.rows.map(row => row.join(", ")),
        datasets: [{
            label: datasetLabel,
            data: result.rows,
            backgroundColor: "rgba(75, 192, 192, 0.5)",
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the rows with filled missing values into Chart.js format.
 *
 * @param {Object} result - The backend response with `rows` (rows with missing values filled).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformFilledMissingToChartJS(result, datasetLabel = "Dataset with Filled Values") {
    return {
        labels: result.rows.map(row => row.join(", ")),
        datasets: [{
            label: datasetLabel,
            data: result.rows,
            backgroundColor: "rgba(54, 162, 235, 0.5)",
            borderColor: "rgba(54, 162, 235, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the rows after log transformation into Chart.js format.
 *
 * @param {Object} result - The backend response with `rows` (log-transformed values).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformLogTransformedToChartJS(result, datasetLabel = "Log-Transformed Dataset") {
    return {
        labels: result.rows.map(row => row[0]),
        datasets: [{
            label: datasetLabel,
            data: result.rows,
            backgroundColor: "rgba(153, 102, 255, 0.5)",
            borderColor: "rgba(153, 102, 255, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the rows after column normalization into Chart.js format.
 *
 * @param {Object} result - The backend response with `rows` (normalized values).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformNormalizedColumnToChartJS(result, datasetLabel = "Normalized Dataset") {
    return {
        labels: result.rows.map(row => row[0]),
        datasets: [{
            label: datasetLabel,
            data: result.rows,
            backgroundColor: "rgba(255, 159, 64, 0.5)",
            borderColor: "rgba(255, 159, 64, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Converts a StandardizeColumn API response (map of column name -> array of numbers)
 * into Chart.js compatible format.
 * 
 * Example input:
 * {
 *   "colA": [0.1, -1.3, 1.2],
 *   "colB": [-0.4, 0.0, 0.4]
 * }
 * 
 * Output format:
 * {
 *   labels: [0, 1, 2],
 *   datasets: [
 *     {
 *       label: "colA",
 *       data: [0.1, -1.3, 1.2],
 *       fill: false,
 *       borderColor: someColor,
 *     },
 *     {
 *       label: "colB",
 *       data: [-0.4, 0.0, 0.4],
 *       fill: false,
 *       borderColor: anotherColor,
 *     }
 *   ]
 * }
 */
function formatStandardizeColumnForChart(data) {
    const labels = Array.from(
      { length: Object.values(data)[0]?.length || 0 },
      (_, i) => i
    );
  
    const colors = [
      "#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"
    ];
  
    const datasets = Object.entries(data).map(([colName, values], idx) => ({
      label: colName,
      data: values,
      fill: false,
      borderColor: colors[idx % colors.length],
      tension: 0.1
    }));
  
    return { labels, datasets };
  }
  
function formatPearsonForChart(data) {
    return {
      labels: ["Pearson Correlation"],
      datasets: [
        {
          label: "Pearson Correlation",
          data: [data.pearson],
          backgroundColor: "rgba(54, 162, 235, 0.5)",
          borderColor: "rgba(54, 162, 235, 1)",
          borderWidth: 1,
        },
      ],
    };
  }

function formatSpearmanForChart(data) {
    return {
      labels: ["Spearman Correlation"],
      datasets: [
        {
          label: "Spearman Correlation",
          data: [data.spearman],
          backgroundColor: "rgba(255, 99, 132, 0.5)",
          borderColor: "rgba(255, 99, 132, 1)",
          borderWidth: 1,
        },
      ],
    };
  }
  
function formatCorrelationMatrixForChart(data) {
    const rowLabels = Object.keys(data);
    const columnSet = new Set();
  
    // Collect all unique column keys
    rowLabels.forEach((row) => {
      const columns = data[row];
      Object.keys(columns).forEach((col) => columnSet.add(col));
    });
  
    const columnLabels = Array.from(columnSet).sort(); // sorted for consistency
  
    const datasets = columnLabels.map((col) => {
      const rowData = rowLabels.map((row) => {
        const value = data[row][col];
        return value !== undefined ? value : null;
      });
  
      return {
        label: col,
        data: rowData,
        backgroundColor: randomRGBA(), // Optional: assign a unique color per column
        borderColor: "rgba(0,0,0,0.1)",
        borderWidth: 1,
      };
    });
  
    return {
      labels: rowLabels,
      datasets,
    };
  }
  
/**
 * Transforms the mean result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "mean": 20 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformMeanToChartJS(result, datasetLabel = "Mean") {
    return {
        labels: ["Mean"],
        datasets: [{
            label: datasetLabel,
            data: [result.mean],
            backgroundColor: "rgba(54, 162, 235, 0.5)",
            borderColor: "rgba(54, 162, 235, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the median result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "median": 30 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformMedianToChartJS(result, datasetLabel = "Median") {
    return {
        labels: ["Median"],
        datasets: [{
            label: datasetLabel,
            data: [result.median],
            backgroundColor: "rgba(75, 192, 192, 0.5)",
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the mode result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "mode": "red" }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformModeToChartJS(result, datasetLabel = "Mode") {
    return {
        labels: ["Mode"],
        datasets: [{
            label: datasetLabel,
            data: [result.mode],
            backgroundColor: "rgba(153, 102, 255, 0.5)",
            borderColor: "rgba(153, 102, 255, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the standard deviation result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "stddev": 5 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformStdDevToChartJS(result, datasetLabel = "Standard Deviation") {
    return {
        labels: ["Std Dev"],
        datasets: [{
            label: datasetLabel,
            data: [result.stddev],
            backgroundColor: "rgba(255, 159, 64, 0.5)",
            borderColor: "rgba(255, 159, 64, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the variance result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "variance": 100 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformVarianceToChartJS(result, datasetLabel = "Variance") {
    return {
        labels: ["Variance"],
        datasets: [{
            label: datasetLabel,
            data: [result.variance],
            backgroundColor: "rgba(153, 255, 51, 0.5)",
            borderColor: "rgba(153, 255, 51, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the min result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "min": 70 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformMinToChartJS(result, datasetLabel = "Min") {
    return {
        labels: ["Min"],
        datasets: [{
            label: datasetLabel,
            data: [result.min],
            backgroundColor: "rgba(255, 99, 132, 0.5)",
            borderColor: "rgba(255, 99, 132, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the max result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "max": 100 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformMaxToChartJS(result, datasetLabel = "Max") {
    return {
        labels: ["Max"],
        datasets: [{
            label: datasetLabel,
            data: [result.max],
            backgroundColor: "rgba(54, 162, 235, 0.5)",
            borderColor: "rgba(54, 162, 235, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the range result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "range": 40 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformRangeToChartJS(result, datasetLabel = "Range") {
    return {
        labels: ["Range"],
        datasets: [{
            label: datasetLabel,
            data: [result.range],
            backgroundColor: "rgba(255, 159, 64, 0.5)",
            borderColor: "rgba(255, 159, 64, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the sum result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "sum": 60 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformSumToChartJS(result, datasetLabel = "Sum") {
    return {
        labels: ["Sum"],
        datasets: [{
            label: datasetLabel,
            data: [result.sum],
            backgroundColor: "rgba(75, 192, 192, 0.5)",
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the count result into Chart.js format.
 *
 * @param {Object} result - The backend response (e.g., { "count": 3 }).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformCountToChartJS(result, datasetLabel = "Count") {
    return {
        labels: ["Count"],
        datasets: [{
            label: datasetLabel,
            data: [result.count],
            backgroundColor: "rgba(153, 102, 255, 0.5)",
            borderColor: "rgba(153, 102, 255, 1)",
            borderWidth: 1,
        }]
    };
}

/**
 * Transforms the histogram result into Chart.js format.
 *
 * @param {Object} result - The backend response with `labels` and `counts`.
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformHistogramToChartJS(result, datasetLabel = "Histogram") {
    return {
        labels: result.labels,
        datasets: [{
            label: datasetLabel,
            data: result.counts,
            backgroundColor: "rgba(75, 192, 192, 0.5)",
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1,
            borderSkipped: false
        }]
    };
}

/**
 * Transforms the KDE result into Chart.js format.
 *
 * @param {Object} result - The backend response with `labels` (x-values) and `densities` (y-values).
 * @param {String} datasetLabel - The label for the dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformKDEToChartJS(result, datasetLabel = "KDE") {
    return {
        labels: result.labels,
        datasets: [{
            label: datasetLabel,
            data: result.densities,
            backgroundColor: "rgba(54, 162, 235, 0.5)",
            borderColor: "rgba(54, 162, 235, 1)",
            borderWidth: 1,
            fill: true
        }]
    };
}

/**
 * Transforms the filtered and sorted data for Chart.js visualization.
 *
 * @param {Object} result - The backend response with filtered and sorted `data`.
 * @param {Array} headers - The headers for the dataset (e.g., ["name", "age"]).
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformFilteredSortedDataToChartJS(result, headers) {
    const labels = result.data.map(row => row[headers[0]]);
    const ages = result.data.map(row => row[headers[1]]);
    
    return {
        labels: labels,
        datasets: [{
            label: "Filtered and Sorted Data",
            data: ages,
            backgroundColor: "rgba(75, 192, 192, 0.5)",
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the Z-score outlier indices into Chart.js format.
 *
 * @param {Object} result - The backend response with outlier `indices`.
 * @param {Array} data - The original dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformZScoreOutliersToChartJS(result, data) {
    const outlierIndices = result.indices;
    const outliers = outlierIndices.map(index => data[index]);

    return {
        labels: data,
        datasets: [{
            label: "Data with Outliers",
            data: data,
            backgroundColor: data.map((_, index) => 
                outlierIndices.includes(index) ? "rgba(255, 99, 132, 0.6)" : "rgba(75, 192, 192, 0.5)"
            ),
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the IQR outlier indices into Chart.js format.
 *
 * @param {Object} result - The backend response with outlier `indices`.
 * @param {Array} data - The original dataset.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformIQROutliersToChartJS(result, data) {
    const outlierIndices = result.indices;
    const outliers = outlierIndices.map(index => data[index]);

    return {
        labels: data,
        datasets: [{
            label: "Data with Outliers",
            data: data,
            backgroundColor: data.map((_, index) => 
                outlierIndices.includes(index) ? "rgba(255, 99, 132, 0.6)" : "rgba(75, 192, 192, 0.5)"
            ),
            borderColor: "rgba(75, 192, 192, 1)",
            borderWidth: 1
        }]
    };
}

/**
 * Transforms the box plot data into Chart.js format.
 *
 * @param {Object} result - The backend response with `labels` and `values` for the box plot.
 * @returns {Object} - An object with `labels` and `datasets` for Chart.js.
 */
function transformBoxPlotDataToChartJS(result) {
    const { labels, values } = result;

    return {
        labels: labels,
        datasets: [{
            label: "Box Plot",
            data: values,
            backgroundColor: "rgba(153, 102, 255, 0.5)",
            borderColor: "rgba(153, 102, 255, 1)",
            borderWidth: 1
        }]
    };
}

export {
    transformGroupedDataToChartJS,
    transformPivotDataToChartJS,
    transformDroppedRowsToChartJS,
    transformFilledMissingToChartJS,
    transformLogTransformedToChartJS,
    transformNormalizedColumnToChartJS,
    formatStandardizeColumnForChart,
    formatPearsonForChart,
    formatSpearmanForChart,
    formatCorrelationMatrixForChart,
    transformMeanToChartJS,
    transformMedianToChartJS,
    transformModeToChartJS,
    transformStdDevToChartJS,
    transformVarianceToChartJS,
    transformMinToChartJS,
    transformMaxToChartJS,
    transformRangeToChartJS,
    transformSumToChartJS,
    transformCountToChartJS,
    transformHistogramToChartJS,
    transformKDEToChartJS,
    transformFilteredSortedDataToChartJS,
    transformZScoreOutliersToChartJS,
    transformIQROutliersToChartJS,
    transformBoxPlotDataToChartJS
  };  
