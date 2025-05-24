import React, { useState, useEffect } from "react";
import Button from "./components/Button";
import Modal from "./components/Modal";
import AuthForm from "./components/AuthForm";
import UploadForm from "./components/UploadForm";
import AccountMenu from "./components/AccountMenu";
import * as styles from "./App.module.css";
import {
  transformGroupedDataToChartJS,
  transformPivotDataToChartJS,
  transformDroppedRowsToChartJS,
  transformFilledMissingToChartJS,
  transformLogTransformedToChartJS,
  transformNormalizedColumnToChartJS,
  transformStandardizeColumnForChart,
  transformPearsonForChart,
  transformSpearmanForChart,
  transformCorrelationMatrixForChart,
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
  transformBoxPlotDataToChartJS,
} from "./helpers/chartDataTransform";
import DatasetChart from "./components/charts/DatasetChart";
import { chartTypeByMethod } from "./components/charts/chartMethodMap";
import DropdownMenu from "./components/DropdownMenu";
import { Bar } from 'react-chartjs-2';
import {
  BoxPlotController,
  BoxAndWiskers,
} from '@sgratzl/chartjs-chart-boxplot';
import { Chart } from 'react-chartjs-2';
import { Line } from 'react-chartjs-2';
import { Chart as ChartJS, CategoryScale, LinearScale, Tooltip, Title } from 'chart.js';
import { MatrixController, MatrixElement } from 'chartjs-chart-matrix';

import {
  BarElement,
  LineElement,
  PointElement,
  ArcElement,
  TimeScale,
  Legend,
} from 'chart.js';

ChartJS.register(
  BarElement,
  LineElement,
  PointElement,
  ArcElement,
  TimeScale,
  Legend,
);

const normalizeMethodKey = (method) => {
  if (!method) return '';

  if (method.startsWith('zscore') || method.startsWith('iqr')) {
    return method.replace('-outliers', '');
  }

  if (method.startsWith('grouped-')) {
    return 'grouped';
  }

  if (method.startsWith('pivot-')) {
    return 'pivot'; 
  }

  if (method.endsWith('-correlation')) {
    return method.replace('-correlation', '');
  }

  if (method === 'correlation-matrix') {
    return 'correlationmatrix';
  }

  if (method.includes('drop-rows-with-missing')) {
    return 'droprowswithmissing';
  }

  if (method.includes('fill-missing-with')) {
    return 'fillmissingwith';
  }

  if (method.includes('apply-log-transformation')) {
    return 'applylogtransformation';
  }

  if (method.includes('normalize-column')) {
   return 'normalizecolumn';
  }

  if (method.includes('standardize-column')) {
    return 'standardizecolumn';
  }

  if (method.includes('filter-sort')) {
    return 'filtersort';
  }

  return method;
};

export default function App() {
  const [modalType, setModalType] = useState(null);
  const [user, setUser] = useState(undefined); // Track the logged-in user
  const [activeDataset, setActiveDataset] = useState(null);
  const [method, setMethod] = useState(null);
  const [labels, setLabels] = useState([]);
  const [datasets, setDatasets] = useState([]);
  const [title, setTitle] = useState("");

  const isUserLoaded = user !== undefined;

  const closeModal = () => setModalType(null);

    useEffect(() => {
      // On mount, try fetching the logged-in user profile
      const fetchUserProfile = async () => {
        try {
          const res = await fetch("http://localhost:8080/user/profile", {
            method: "GET",
            credentials: "include",
          });
          if (res.status === 401) {
            // Not logged in — expected
            console.debug("User not logged in yet");
            setUser(null);
            return;
          }

          if (!res.ok) {
            throw new Error("Unexpected error fetching user profile");
          }

          const data = await res.json();
          setUser(data);
        } catch (err) {
          console.error("Error fetching user profile:", err.message);
        }
      };

      fetchUserProfile();
    }, []);

  const handleUploadSuccess = async (datasetInfo) => {
    console.log("Upload success data:", datasetInfo);
    const id = datasetInfo.ID || datasetInfo.id;

    if (!id) {
      setActiveDataset(null);
      closeModal();
      return;
    }

    try {
      // Step 1: Fetch the full dataset (with rows) from backend
      const res = await fetch(`http://localhost:8080/datasets/${id}`, {
        credentials: "include",
      });

      if (!res.ok) throw new Error("Failed to fetch full dataset");

      const fullDataset = await res.json();

      // Step 2: Update state with full dataset including .rows
      console.log("Upload fullDataset:", fullDataset);
      setActiveDataset(fullDataset);
    } catch (err) {
      console.error("Error loading full dataset after upload:", err);
      alert("Upload succeeded but failed to load dataset rows.");
      setActiveDataset({ ...datasetInfo, id }); // fallback to metadata only
    }

    // Step 3: Close modal
    setTimeout(() => {
      closeModal();
    }, 50);
  };

  const handleDatasetDelete = async () => {
    console.log(activeDataset);
    console.log("Calling delete for", activeDataset?.id);
    if (!activeDataset?.id) {
      console.log("No active dataset/ dataset_ID found");
      return;
    }
    console.log(activeDataset);
    console.log("Calling delete for", activeDataset?.id);

    const res = await fetch(`http://localhost:8080/datasets/${activeDataset.id}`, {
      method: "DELETE",
      credentials: "include",
    });

    if (res.ok) {
      setActiveDataset(null);
      setTimeout(() => {
        closeModal();
      }, 5);
    } else {
      const err = await res.json().catch(() => ({}));
      alert(err.error || "Delete failed");
    }
  };

  const handleLogout = async () => {
    try {
      await fetch("http://localhost:8080/user/logout", {
        method: "POST",
        credentials: "include",
      });
      setUser(null);
    } catch (error) {
      console.error("Logout failed", error);
    }
  };

    const requiresColumn = (method) => {
      const methodsNeedingColumns = new Set([
        "mean", "median", "mode", "stddev", "min", "max", "variance", "range",
        "sum", "histogram", "kde", "zscore-outliers", "iqr-outliers",
        "boxplot", "normalize-column", "standardize-column",
        "apply-log-transformation", "grouped-sum", "grouped-mean", "grouped-min",
        "grouped-max", "grouped-median", "grouped-stddev", "pivot-sum", "pivot-mean",
        "pivot-min", "pivot-max", "pivot-median", "pivot-stddev", "pivot-count"
      ]);
      return methodsNeedingColumns.has(method);
    };

    const requiresGroup = (method) => {
      const methodsNeedingGroup = new Set([
        "grouped-sum", "grouped-mean", "grouped-min",
        "grouped-max", "grouped-median", "grouped-stddev",
        "grouped-count"
      ]);
      return methodsNeedingGroup.has(method);
    }

    const requiresRow = (method) => {
      const methodsNeedingGroup = new Set([
        "pivot-sum", "pivot-mean", "pivot-min",
        "pivot-max", "pivot-median", "pivot-stddev",
        "pivot-count"
      ]);
      return methodsNeedingGroup.has(method);
    }

    const requiresValue = (method) => {
      const methodsNeedingGroup = new Set([
        "pivot-sum", "pivot-mean", "pivot-min",
        "pivot-max", "pivot-median", "pivot-stddev",
        "pivot-count"
      ]);
      return methodsNeedingGroup.has(method);
    }

    const requiresXandY = (method) => {
      const methodsNeedingXandY = new Set ([
        "pearson-correlation", "spearman-correlation"
      ])
    return methodsNeedingXandY.has(method);
  }

  const requiresMethod = (method) => {
      const methodsNeedingXandY = new Set ([
        "correlation-matrix"
      ])
    return methodsNeedingXandY.has(method);
  }

  function requiresColumnsArray(method) {
    return ["drop-rows-with", "fill-missing-with"].includes(method);
  }

  function requiresMultiColumn(method) {
    return ["correlation-matrix"].includes(method);
  }

  const handleSelect = async (group, method, extraMethod = null) => {
    if (!activeDataset?.id) return;

    const defaultColumn = Array.isArray(activeDataset.columns) && activeDataset.columns.length > 0
      ? activeDataset.columns[6]
      : null;

    if (!defaultColumn) {
      console.warn("No valid default column selected.");
      return;
    }

    const defaultGroupBy = Array.isArray(activeDataset.columns) && activeDataset.columns.length > 0
        ? activeDataset.columns[2]
        : null;

    if (!defaultGroupBy) {
      console.warn("No valid default group_by column selected.");
      return;
    }

    const defaultRow = Array.isArray(activeDataset.columns) && activeDataset.columns.length > 0
        ? activeDataset.columns[1] // "Age"
        : null;

    if (!defaultRow) {
      console.warn("No valid default row selected.");
      return;
    }

    const defaultValue = Array.isArray(activeDataset.columns) && activeDataset.columns.length > 0
        ? activeDataset.columns[3] // "Income"
        : null;

    if (!defaultValue) {
      console.warn("No valid default value selected.");
      return;
    }

    console.log("Rows:", activeDataset?.rows);
    console.log("First row:", activeDataset?.rows?.[0]);

    const colIndex = activeDataset.columns.findIndex(c => c === defaultColumn);
    const columnData = activeDataset.rows.map(row => row[colIndex]);
  
    console.log("Active Dataset:", activeDataset);

    let url = `http://localhost:8080/analytics/${group}/${method}?dataset_id=${activeDataset?.id}`;

    if (requiresGroup(method)) {
      url += `&group_by=${encodeURIComponent(defaultGroupBy)}`;
    }

    if (requiresRow(method)) {
      url += `&row_field=${encodeURIComponent(defaultRow)}`;
    }

    if (requiresMethod(method)) {
      url += `&method=${encodeURIComponent(extraMethod)}`
    }

    if (requiresColumn(method)) {
      url += `&column=${encodeURIComponent(defaultColumn)}`;
    }

    if (requiresMultiColumn(method)) {
      url += `&column=${encodeURIComponent(defaultColumn)}&column=${encodeURIComponent(defaultValue)}`;
    }

    if (requiresXandY(method)) {
      url += `&row_field=${encodeURIComponent(defaultRow)}&column=${encodeURIComponent(defaultColumn)}`;
    }

    if (requiresValue(method)) {
      url += `&value_field=${encodeURIComponent(defaultValue)}`;
    }

    console.log("Requesting:", url);

    console.log("defaultColumn:", defaultColumn);
    console.log("defaultGroupBy:", defaultGroupBy);
    console.log("columnData:", columnData);

    try {
      const res = await fetch(url, {
        method: "GET",
        credentials: "include",
      });
      const raw = await res.json();

      console.log("raw:", raw);

      let transformed;
      switch (method) {
        case 'median':
          transformed = transformMedianToChartJS(raw);
          break;
        case 'mean':
          transformed = transformMeanToChartJS(raw);
          break;
        case 'mode':
          transformed = transformModeToChartJS(raw);
          break;
        case 'variance':
          transformed = transformVarianceToChartJS(raw);
          break;
        case 'min':
          transformed = transformMinToChartJS(raw);
          break;
        case 'max':
          transformed = transformMaxToChartJS(raw);
          break;
        case 'range':
          transformed = transformRangeToChartJS(raw);
          break;
        case 'sum':
          transformed = transformSumToChartJS(raw);
          break;
        case 'count':
          transformed = transformCountToChartJS(raw);
          break;
        case 'stddev':
          transformed = transformStdDevToChartJS(raw);
          break;
        case 'histogram':
          transformed = transformHistogramToChartJS(raw);
          break;
        case 'kde':
          transformed = transformKDEToChartJS(raw);
          break;
        case 'filtersort':
          transformed = transformFilteredSortedDataToChartJS(raw);
          break;
        case 'zscore-outliers':
          transformed = transformZScoreOutliersToChartJS(raw, columnData);
          break;
        case 'iqr-outliers':
          transformed = transformIQROutliersToChartJS(raw, columnData);
          break;
        case 'boxplot':
          transformed = transformBoxPlotDataToChartJS(raw);
          break;
        case 'grouped-sum':
        case 'grouped-mean':
        case 'grouped-count':
        case 'grouped-min':
        case 'grouped-max':
        case 'grouped-median':
        case 'grouped-stddev':
          transformed = transformGroupedDataToChartJS(raw.results, "Result");
          break;
        case 'pivot-sum':
        case 'pivot-mean':
        case 'pivot-count':
        case 'pivot-min':
        case 'pivot-max':
        case 'pivot-median':
        case 'pivot-stddev':
          transformed = transformPivotDataToChartJS(raw.results);
          break;
        case 'drop-rows-with-missing':
          transformed = transformDroppedRowsToChartJS(raw);
          break;
        case 'fill-missing-with':
          transformed = transformFilledMissingToChartJS(raw);
          break;
        case 'apply-log-transformation':
          transformed = transformLogTransformedToChartJS(raw);
          break;
        case 'normalize-column':
          transformed = transformNormalizedColumnToChartJS(raw);
          break;
        case 'standardize-column':
          transformed = transformStandardizeColumnForChart(raw);
          break;
        case 'pearson-correlation':
          transformed = transformPearsonForChart(raw.results);
          break;
        case 'spearman-correlation':
          transformed = transformSpearmanForChart(raw.results);
          break;
        case 'correlation-matrix':
          transformed = transformCorrelationMatrixForChart(raw.results);
          console.log("Matrix response:", transformed);
          break;
        default:
          console.error(`No transform defined for method: ${method}`);
          throw new Error(`Unsupported analysis method: ${method}`);
      }
      const normalizedMethod = normalizeMethodKey(method);
      console.log(`Chart type for method "${method}" (normalized: "${normalizedMethod}") → ${chartTypeByMethod[normalizedMethod]}`);

      // Update chart state
      setLabels(transformed.labels);
      setDatasets(transformed.datasets);
      setTitle(transformed.title);
      setMethod(method);
    } catch (err) {
      console.error('API error:', err);
    }
  };

  return (
    <div className={styles.app}>
      {/* Header */}
      <header className={styles.header}>
        <div className={styles.logoContainer}>
          <img
            src="/assets/InsightForge_Logo.png"
            alt="InsightForge Logo"
            className={styles.logo}
          />
        </div>

        <div className={styles.titles}>
          <h1 className={styles.title}>InsightForge</h1>
          <h2 className={styles.title2}>Visualize your data, forge your path</h2>
        </div>

        {isUserLoaded && (
          <div className={styles.buttonContainer}>
            {!user ? (
              <Button text="Login / Register" onClick={() => setModalType("auth")} />
            ) : (
              <div className={styles.accountMenu}>
                <AccountMenu user={user} onLogout={handleLogout} />
              </div>
            )}
            <Button
              text="Upload/Delete Dataset"
              onClick={() => setModalType("upload")}
              disabled={!user}
              title={!user ? "Login to upload a dataset" : ""}
            />
          </div>
        )}
      </header>

      <div className={styles.divider}></div>

      {/* Content with Sidebar */}
      <div className={styles.contentWrapper}>
        <aside className={styles.sidebar}>
          <h3>Forge Tools</h3>

          <DropdownMenu
            label="Aggregation"
            disabled={!activeDataset}
            options={[
              {
                text: "Grouped Data",
                children: [
                  { text: "Grouped By", onClick: () => handleSelect("aggregation", "grouped-by") },
                  { text: "Grouped Sum", onClick: () => handleSelect("aggregation", "grouped-sum") },
                  { text: "Grouped Mean", onClick: () => handleSelect("aggregation", "grouped-mean") },
                  { text: "Grouped Count", onClick: () => handleSelect("aggregation", "grouped-count") },
                  { text: "Grouped Min", onClick: () => handleSelect("aggregation", "grouped-min") },
                  { text: "Grouped Max", onClick: () => handleSelect("aggregation", "grouped-max") },
                  { text: "Grouped Median", onClick: () => handleSelect("aggregation", "grouped-median") },
                  { text: "Grouped StdDev", onClick: () => handleSelect("aggregation", "grouped-stddev") }
                ]
              },
              {
                text: "Pivot Data",
                children: [
                  { text: "Pivot Sum", onClick: () => handleSelect("aggregation", "pivot-sum") },
                  { text: "Pivot Mean", onClick: () => handleSelect("aggregation", "pivot-mean") },
                  { text: "Pivot Count", onClick: () => handleSelect("aggregation", "pivot-count") },
                  { text: "Pivot Min", onClick: () => handleSelect("aggregation", "pivot-min") },
                  { text: "Pivot Max", onClick: () => handleSelect("aggregation", "pivot-max") },
                  { text: "Pivot Median", onClick: () => handleSelect("aggregation", "pivot-median") },
                  { text: "Pivot StdDev", onClick: () => handleSelect("aggregation", "pivot-stddev") }
                ]
              }
            ]}
          />

          <DropdownMenu
              label="Cleaning"
              disabled={!activeDataset}
              options={[
                { text: "Dropped Rows", onClick: () => handleSelect("cleaning", "drop-rows-with-missing") },
                { text: "Fill Missing With", onClick: () => handleSelect("cleaning", "fill-missing-with") },
                { text: "Log Transform", onClick: () => handleSelect("cleaning", "apply-log-transformation") },
                { text: "Normalize Column", onClick: () => handleSelect("cleaning", "normalize-column") },
                { text: "Standardize Column", onClick: () => handleSelect("cleaning", "standardize-column") },
              ]}
            />

          <DropdownMenu
            label="Correlation"
            disabled={!activeDataset}
            options={[
              { text: "Pearson", onClick: () => handleSelect("correlation", "pearson-correlation") },
              { text: "Spearman", onClick: () => handleSelect("correlation", "spearman-correlation") },
              {
                text: "Correlation Matrix",
                children: [
                  {
                    text: "Pearson Matrix",
                    onClick: () => handleSelect("correlation", "correlation-matrix", "pearson"),
                  },
                  {
                    text: "Spearman Matrix",
                    onClick: () => handleSelect("correlation", "correlation-matrix", "spearman"),
                  },
                ],
              },
            ]}
          />

          <DropdownMenu
            label="Descriptive Statistics"
            disabled={!activeDataset}
            options={[
              { text: "Mean", onClick: () => handleSelect("descriptives", "mean") },
              { text: "Median", onClick: () => handleSelect("descriptives", "median") },
              { text: "Mode", onClick: () => handleSelect("descriptives", "mode") },
              { text: "Standard Deviation", onClick: () => handleSelect("descriptives", "stddev") },
              { text: "Variance", onClick: () => handleSelect("descriptives", "variance") },
              { text: "Min", onClick: () => handleSelect("descriptives", "min") },
              { text: "Max", onClick: () => handleSelect("descriptives", "max") },
              { text: "Sum", onClick: () => handleSelect("descriptives", "sum") },
              { text: "Range", onClick: () => handleSelect("descriptives", "range") },
              { text: "Count", onClick: () => handleSelect("descriptives", "count") },
            ]}
          />

          <DropdownMenu
            label="Distribution"
            disabled={!activeDataset}
            options={[
              { text: "Histogram", onClick: () => handleSelect("distribution", "histogram") },
              { text: "KDE", onClick: () => handleSelect("distribution", "kde") },
            ]}
          />

          <DropdownMenu
            label="Filter Sort"
            disabled={!activeDataset}
            options={[
              { text: "Filter Sort", onClick: () => handleSelect("filtersort", "filter-sort") },
            ]}
          />

          <DropdownMenu
            label="Outliers"
            disabled={!activeDataset}
            options={[
              { text: "Z Score", onClick: () => handleSelect("outliers", "zscore-outliers") },
              { text: "IQR", onClick: () => handleSelect("outliers", "iqr-outliers") },
              { text: "Box Plot", onClick: () => handleSelect("outliers", "boxplot") },
            ]}
          />
        </aside>

        <main className={styles.mainContent}>
          <div className={styles.mainText}>Explore your datasets here.</div>

           <div className={styles.chartContainer}>
              {datasets.length > 0 && method ? (
                <DatasetChart
                  chartType={chartTypeByMethod[normalizeMethodKey(method)]}
                  labels={labels}
                  datasets={datasets}
                  title={title}
                />
              ) : (
                <p className="text-muted">Upload a dataset and select a method to visualize it.</p>
              )}
            </div>

          {modalType === "auth" && (
            <Modal title="Login / Register" onClose={closeModal}>
              <AuthForm
                onSuccess={(userData) => {
                  setUser(userData);
                  closeModal();
                }}
              />
            </Modal>
          )}

          {modalType === "upload" && (
            <Modal onClose={closeModal}>
              <UploadForm
                user={user}
                datasetUploaded={!!activeDataset}
                datasetID={activeDataset?.id}
                onSuccess={handleUploadSuccess}
                onDelete={handleDatasetDelete}
              />
            </Modal>
          )}
        </main>
      </div>

      {/* Footer */}
      <footer className={styles.footer}>
        © {new Date().getFullYear()} InsightForge. All rights reserved.
      </footer>
    </div>
  );
}
