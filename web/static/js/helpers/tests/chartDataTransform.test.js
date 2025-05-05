import {
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
  } from '../chartDataTransform';
  

describe("ChartJS Helpers", () => {
    test("transformGroupedDataToChartJS", () => {
        const data = { "GroupA": 123, "GroupB": 456 };
        const result = transformGroupedDataToChartJS(data, "Test Label");

        expect(result.labels).toEqual(["GroupA", "GroupB"]);
        expect(result.datasets[0].label).toBe("Test Label");
        expect(result.datasets[0].data).toEqual([123, 456]);
    });

    test("transformPivotDataToChartJS", () => {
        const pivotTable = {
            "RowA": { "Col1": 10, "Col2": 20 },
            "RowB": { "Col1": 15, "Col2": 25 }
        };
        const result = transformPivotDataToChartJS(pivotTable);

        expect(result.labels).toEqual(["RowA", "RowB"]);
        expect(result.datasets.length).toBe(2);
        expect(result.datasets[0].label).toBe("Col1");
        expect(result.datasets[1].label).toBe("Col2");
        expect(result.datasets[0].data).toEqual([10, 15]);
        expect(result.datasets[1].data).toEqual([20, 25]);
    });

    test("transformDroppedRowsToChartJS", () => {
        const result = { rows: [["Alice", 24], ["Bob", 30]] };
        const datasetLabel = "Cleaned Dataset";
        const chartData = transformDroppedRowsToChartJS(result, datasetLabel);

        expect(chartData.labels).toEqual(["Alice, 24", "Bob, 30"]);
        expect(chartData.datasets[0].label).toBe(datasetLabel);
        expect(chartData.datasets[0].data).toEqual([["Alice", 24], ["Bob", 30]]);
    });

    test("transformFilledMissingToChartJS", () => {
        const result = { rows: [["Alice", 24], ["Bob", 30]] };
        const datasetLabel = "Dataset with Filled Values";
        const chartData = transformFilledMissingToChartJS(result, datasetLabel);

        expect(chartData.labels).toEqual(["Alice, 24", "Bob, 30"]);
        expect(chartData.datasets[0].label).toBe(datasetLabel);
        expect(chartData.datasets[0].data).toEqual([["Alice", 24], ["Bob", 30]]);
    });

    test("transformLogTransformedToChartJS", () => {
        const result = { rows: [[1, 2], [2, 3], [3, 4]] };
        const datasetLabel = "Log-Transformed Dataset";
        const chartData = transformLogTransformedToChartJS(result, datasetLabel);

        expect(chartData.labels).toEqual([1, 2, 3]);
        expect(chartData.datasets[0].label).toBe(datasetLabel);
        expect(chartData.datasets[0].data).toEqual([[1, 2], [2, 3], [3, 4]]);
    });

    test("transformNormalizedColumnToChartJS", () => {
        const result = { rows: [[0.1, 0.2], [0.3, 0.4]] };
        const datasetLabel = "Normalized Dataset";
        const chartData = transformNormalizedColumnToChartJS(result, datasetLabel);

        expect(chartData.labels).toEqual([0.1, 0.3]);
        expect(chartData.datasets[0].label).toBe(datasetLabel);
        expect(chartData.datasets[0].data).toEqual([[0.1, 0.2], [0.3, 0.4]]);
    });

    test("formatStandardizeColumnForChart", () => {
        const data = {
            "colA": [0.1, -1.3, 1.2],
            "colB": [-0.4, 0.0, 0.4]
        };
        const result = formatStandardizeColumnForChart(data);

        expect(result.labels).toEqual([0, 1, 2]);
        expect(result.datasets.length).toBe(2);
        expect(result.datasets[0].label).toBe("colA");
        expect(result.datasets[1].label).toBe("colB");
        expect(result.datasets[0].data).toEqual([0.1, -1.3, 1.2]);
        expect(result.datasets[1].data).toEqual([-0.4, 0.0, 0.4]);
    });

    test('formatPearsonForChart should return valid chart format for Pearson correlation', () => {
        const data = { pearson: 0.8 };
        const result = formatPearsonForChart(data);
    
        expect(result).toEqual({
            labels: ["Pearson Correlation"],
            datasets: [
                {
                    label: "Pearson Correlation",
                    data: [0.8],
                    backgroundColor: "rgba(54, 162, 235, 0.5)",
                    borderColor: "rgba(54, 162, 235, 1)",
                    borderWidth: 1,
                },
            ],
        });
    });

    test('formatSpearmanForChart should return valid chart format for Spearman correlation', () => {
        const data = { spearman: 0.9 };
        const result = formatSpearmanForChart(data);
    
        expect(result).toEqual({
            labels: ["Spearman Correlation"],
            datasets: [
                {
                    label: "Spearman Correlation",
                    data: [0.9],
                    backgroundColor: "rgba(255, 99, 132, 0.5)",
                    borderColor: "rgba(255, 99, 132, 1)",
                    borderWidth: 1,
                },
            ],
        });
    });

    test('formatCorrelationMatrixForChart should return valid chart format for correlation matrix', () => {
        const data = {
            A: { B: 0.5, C: 0.7 },
            B: { A: 0.5, C: 0.8 },
            C: { A: 0.7, B: 0.8 },
        };
    
        const result = formatCorrelationMatrixForChart(data);
    
        expect(result).toEqual({
            labels: ["A", "B", "C"],
            datasets: [
                {
                    label: "A",
                    data: [null, 0.5, 0.7],
                    backgroundColor: expect.any(String),
                    borderColor: "rgba(0,0,0,0.1)",
                    borderWidth: 1,
                },
                {
                    label: "B",
                    data: [0.5, null, 0.8],
                    backgroundColor: expect.any(String),
                    borderColor: "rgba(0,0,0,0.1)",
                    borderWidth: 1,
                },
                {
                    label: "C",
                    data: [0.7, 0.8, null],
                    backgroundColor: expect.any(String),
                    borderColor: "rgba(0,0,0,0.1)",
                    borderWidth: 1,
                },
            ],
        });
    });

    test('transformMeanToChartJS should return valid chart format for mean', () => {
        const result = { mean: 20 };
        const chartJSResult = transformMeanToChartJS(result);
    
        expect(chartJSResult).toEqual({
            labels: ["Mean"],
            datasets: [
                {
                    label: "Mean",
                    data: [20],
                    backgroundColor: "rgba(54, 162, 235, 0.5)",
                    borderColor: "rgba(54, 162, 235, 1)",
                    borderWidth: 1,
                },
            ],
        });
    });

    test('transformMedianToChartJS should return valid chart format for median', () => {
        const result = { median: 30 };
        const chartJSResult = transformMedianToChartJS(result);
    
        expect(chartJSResult).toEqual({
            labels: ["Median"],
            datasets: [
                {
                    label: "Median",
                    data: [30],
                    backgroundColor: "rgba(75, 192, 192, 0.5)",
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1,
                },
            ],
        });
    });

    test('transformModeToChartJS should return valid chart format for mode', () => {
        const result = { mode: "red" };
        const chartJSResult = transformModeToChartJS(result);
    
        expect(chartJSResult).toEqual({
            labels: ["Mode"],
            datasets: [
                {
                    label: "Mode",
                    data: ["red"],
                    backgroundColor: "rgba(153, 102, 255, 0.5)",
                    borderColor: "rgba(153, 102, 255, 1)",
                    borderWidth: 1,
                },
            ],
        });
    });
});

describe('Transformation Functions', () => {
    describe('transformStdDevToChartJS', () => {
        it('should transform standard deviation result to Chart.js format', () => {
            const result = { stddev: 5 };
            const datasetLabel = "Standard Deviation";
            const chartData = transformStdDevToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Std Dev"],
                datasets: [{
                    label: datasetLabel,
                    data: [5],
                    backgroundColor: "rgba(255, 159, 64, 0.5)",
                    borderColor: "rgba(255, 159, 64, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformVarianceToChartJS', () => {
        it('should transform variance result to Chart.js format', () => {
            const result = { variance: 100 };
            const datasetLabel = "Variance";
            const chartData = transformVarianceToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Variance"],
                datasets: [{
                    label: datasetLabel,
                    data: [100],
                    backgroundColor: "rgba(153, 255, 51, 0.5)",
                    borderColor: "rgba(153, 255, 51, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformMinToChartJS', () => {
        it('should transform min result to Chart.js format', () => {
            const result = { min: 70 };
            const datasetLabel = "Min";
            const chartData = transformMinToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Min"],
                datasets: [{
                    label: datasetLabel,
                    data: [70],
                    backgroundColor: "rgba(255, 99, 132, 0.5)",
                    borderColor: "rgba(255, 99, 132, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformMaxToChartJS', () => {
        it('should transform max result to Chart.js format', () => {
            const result = { max: 100 };
            const datasetLabel = "Max";
            const chartData = transformMaxToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Max"],
                datasets: [{
                    label: datasetLabel,
                    data: [100],
                    backgroundColor: "rgba(54, 162, 235, 0.5)",
                    borderColor: "rgba(54, 162, 235, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformRangeToChartJS', () => {
        it('should transform range result to Chart.js format', () => {
            const result = { range: 40 };
            const datasetLabel = "Range";
            const chartData = transformRangeToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Range"],
                datasets: [{
                    label: datasetLabel,
                    data: [40],
                    backgroundColor: "rgba(255, 159, 64, 0.5)",
                    borderColor: "rgba(255, 159, 64, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformSumToChartJS', () => {
        it('should transform sum result to Chart.js format', () => {
            const result = { sum: 60 };
            const datasetLabel = "Sum";
            const chartData = transformSumToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Sum"],
                datasets: [{
                    label: datasetLabel,
                    data: [60],
                    backgroundColor: "rgba(75, 192, 192, 0.5)",
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformCountToChartJS', () => {
        it('should transform count result to Chart.js format', () => {
            const result = { count: 3 };
            const datasetLabel = "Count";
            const chartData = transformCountToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["Count"],
                datasets: [{
                    label: datasetLabel,
                    data: [3],
                    backgroundColor: "rgba(153, 102, 255, 0.5)",
                    borderColor: "rgba(153, 102, 255, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformHistogramToChartJS', () => {
        it('should transform histogram result to Chart.js format', () => {
            const result = { labels: ["A", "B", "C"], counts: [5, 10, 15] };
            const datasetLabel = "Histogram";
            const chartData = transformHistogramToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: ["A", "B", "C"],
                datasets: [{
                    label: datasetLabel,
                    data: [5, 10, 15],
                    backgroundColor: "rgba(75, 192, 192, 0.5)",
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1,
                    borderSkipped: false
                }]
            });
        });
    });

    describe('transformKDEToChartJS', () => {
        it('should transform KDE result to Chart.js format', () => {
            const result = { labels: [1, 2, 3], densities: [0.1, 0.5, 0.3] };
            const datasetLabel = "KDE";
            const chartData = transformKDEToChartJS(result, datasetLabel);
            
            expect(chartData).toEqual({
                labels: [1, 2, 3],
                datasets: [{
                    label: datasetLabel,
                    data: [0.1, 0.5, 0.3],
                    backgroundColor: "rgba(54, 162, 235, 0.5)",
                    borderColor: "rgba(54, 162, 235, 1)",
                    borderWidth: 1,
                    fill: true
                }]
            });
        });
    });

    describe('transformFilteredSortedDataToChartJS', () => {
        it('should transform filtered and sorted data to Chart.js format', () => {
            const result = { data: [{ name: "John", age: 25 }, { name: "Jane", age: 30 }] };
            const headers = ["name", "age"];
            const chartData = transformFilteredSortedDataToChartJS(result, headers);
            
            expect(chartData).toEqual({
                labels: ["John", "Jane"],
                datasets: [{
                    label: "Filtered and Sorted Data",
                    data: [25, 30],
                    backgroundColor: "rgba(75, 192, 192, 0.5)",
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformZScoreOutliersToChartJS', () => {
        it('should transform Z-score outlier indices to Chart.js format', () => {
            const result = { indices: [0, 2] };
            const data = [10, 20, 30];
            const chartData = transformZScoreOutliersToChartJS(result, data);
            
            expect(chartData).toEqual({
                labels: [10, 20, 30],
                datasets: [{
                    label: "Data with Outliers",
                    data: [10, 20, 30],
                    backgroundColor: ["rgba(255, 99, 132, 0.6)", "rgba(75, 192, 192, 0.5)", "rgba(255, 99, 132, 0.6)"],
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformIQROutliersToChartJS', () => {
        it('should transform IQR outlier indices to Chart.js format', () => {
            const result = { indices: [1] };
            const data = [10, 20, 30];
            const chartData = transformIQROutliersToChartJS(result, data);
            
            expect(chartData).toEqual({
                labels: [10, 20, 30],
                datasets: [{
                    label: "Data with Outliers",
                    data: [10, 20, 30],
                    backgroundColor: ["rgba(75, 192, 192, 0.5)", "rgba(255, 99, 132, 0.6)", "rgba(75, 192, 192, 0.5)"],
                    borderColor: "rgba(75, 192, 192, 1)",
                    borderWidth: 1
                }]
            });
        });
    });

    describe('transformBoxPlotDataToChartJS', () => {
        it('should transform box plot data to Chart.js format', () => {
            const result = { labels: ["Dataset1"], values: [1, 2, 3, 4, 5] };
            const chartData = transformBoxPlotDataToChartJS(result);
            
            expect(chartData).toEqual({
                labels: ["Dataset1"],
                datasets: [{
                    label: "Box Plot",
                    data: [1, 2, 3, 4, 5],
                    backgroundColor: "rgba(153, 102, 255, 0.5)",
                    borderColor: "rgba(153, 102, 255, 1)",
                    borderWidth: 1
                }]
            });
        });
    });
});
