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
  transformMeanMedianToChartJS,
  transformStdDevVarianceToChartJS,
  transformMinMaxToChartJS,
  transformRangeStdDevToChartJS,
  transformModeMedianToChartJS,
  transformSumCountToChartJS,
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
  const [user, setUser] = useState(null);
  const [activeDataset, setActiveDataset] = useState(null);
  const [method, setMethod] = useState(null);
  const [labels, setLabels] = useState([]);
  const [datasets, setDatasets] = useState([]);
  const [title, setTitle] = useState("");
  const [errorDisabled, setErrorDisabled] = useState(false);

  const [isUserLoaded, setIsUserLoaded] = useState(false);

  const closeModal = () => setModalType(null);

    const fetchUserProfile = async () => {
    try {
      const res = await fetch("http://localhost:8080/user/profile", {
        method: "GET",
        credentials: "include",
      });

      if (res.status === 401) {
        console.debug("User not logged in or session expired");
        setUser(null);
        localStorage.removeItem("user");
        return;
      }

      if (!res.ok) {
        throw new Error(`Server error: ${res.status}`);
      }

      const data = await res.json();
      setUser(data);
      localStorage.setItem("user", JSON.stringify(data));
    } catch (err) {
      console.error("Error fetching user profile:", err.message);
      setUser(null);
      localStorage.removeItem("user");
    } finally {
      setIsUserLoaded(true);
    }
  };

  useEffect(() => {
    // Restore cached dataset
    const stored = localStorage.getItem("activeDataset");
    if (stored) {
      try {
        setActiveDataset(JSON.parse(stored));
      } catch (err) {
        console.error("Failed to parse stored dataset:", err);
      }
    }

    // Restore cached user
    const cached = localStorage.getItem("user");
    if (cached) {
      setUser(JSON.parse(cached));
    }

    // Always revalidate session
    fetchUserProfile();
  }, []);

  useEffect(() => {
    const restoreDataset = async () => {
      const stored = localStorage.getItem("activeDataset");
      if (!stored) return;

      try {
        const parsed = JSON.parse(stored);

        if (activeDataset && activeDataset.id === parsed.id) {
          return;
        }

        const res = await fetch(`http://localhost:8080/datasets/${parsed.id}`, {
          credentials: "include",
        });

        if (!res.ok) {
          throw new Error("Dataset not found on server");
        }

        const fullDataset = await res.json();
        setActiveDataset(fullDataset);
      } catch (err) {
        console.warn("Invalid cached dataset removed:", err.message);
        localStorage.removeItem("activeDataset");
        setActiveDataset(null);
      }
    };

    if (isUserLoaded && !activeDataset) {
      restoreDataset();
    }
  }, [isUserLoaded]);

  useEffect(() => {
    const shouldLogout = sessionStorage.getItem("logoutOnReload") === "true";

    if (shouldLogout) {
      fetch("http://localhost:8080/user/logout", {
        method: "POST",
        credentials: "include",
      }).finally(() => {
        sessionStorage.removeItem("logoutOnReload");
        window.location.reload(); // force refresh to clear state
      });
      return; // Don't continue initializing app state
    }

    // 2. Restore cached user from localStorage
    const cached = localStorage.getItem("user");
    if (cached) {
      setUser(JSON.parse(cached));
    }

    // 3. Restore cached dataset from localStorage
    const stored = localStorage.getItem("activeDataset");
    if (stored) {
      try {
        setActiveDataset(JSON.parse(stored));
      } catch (err) {
        console.error("Failed to parse stored dataset:", err);
      }
    }

    // 4. Always revalidate session
    fetchUserProfile();
  }, []);

    useEffect(() => {
      if (activeDataset) {
        localStorage.setItem("activeDataset", JSON.stringify(activeDataset));
      } else {
        localStorage.removeItem("activeDataset");
      }
    }, [activeDataset]);

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
    if (!activeDataset?.id) {
      console.log("No active dataset / dataset_ID found");
      return;
    }

    try {
      const res = await fetch(`http://localhost:8080/datasets/${activeDataset.id}`, {
        method: "DELETE",
        credentials: "include",
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        alert(err.error || "Delete failed");
        return;
      }

      setActiveDataset(null);
      localStorage.removeItem("activeDataset");

      setTimeout(() => {
        closeModal();
      }, 5);
    } catch (err) {
      console.error("Error deleting dataset:", err);
      alert("Delete failed");
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

  useEffect(() => {
    const token = new URLSearchParams(window.location.search).get('verify_token');

    if (!token) return;

    fetch("http://localhost:8080/user/set-verify-cookie", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ token }),
    })
      .then(() => {
        return fetch("http://localhost:8080/user/verify", {
          method: "POST",
          credentials: "include",
        });
      })
      .then(() => {
        window.history.replaceState(null, "", "/verify-success");
        window.location.href = "/verify-success";
      });
  }, []);

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

    const colIndex = activeDataset.columns.findIndex(c => c === defaultColumn);
    const columnData = activeDataset.rows.map(row => row[colIndex]);

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

    try {
      const res = await fetch(url, {
        method: "GET",
        credentials: "include",
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || "Unknown error");
      }

      setErrorDisabled(false);

      const raw = await res.json();

      let transformed;
      switch (method) {
        case 'mean':
        case 'median': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [meanRes, medianRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/mean?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/median?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
          ]);

          if (!meanRes.ok || !medianRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const meanRaw = await meanRes.json();
          const medianRaw = await medianRes.json();

          // Merge mean and median responses
          const merged = meanRaw.results.map((meanItem, idx) => ({
            label: meanItem.label,
            mean: meanItem.value,
            median: medianRaw.results[idx]?.value,
          }));

          transformed = transformMeanMedianToChartJS(merged);
          break;
        }
        case 'mode':
        case 'median': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [modeRes, medianRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/mode?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/median?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
          ]);

          if (!modeRes.ok || !medianRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const modeRaw = await modeRes.json();
          const medianRaw = await medianRes.json();

          const merged = modeRaw.results.map((modeItem, idx) => ({
            label: modeItem.label,
            mean: modeItem.value,
            median: medianRaw.results[idx]?.value,
          }));

          transformed = transformModeMedianToChartJS(merged);
          break;
        }
        case 'min':
        case 'max': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [minRes, maxRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/min?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/max?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
          ]);

          if (!minRes.ok || !maxRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const minRaw = await minRes.json();
          const maxRaw = await maxRes.json();

          const merged = minRaw.results.map((minItem, idx) => ({
            label: minItem.label,
            mean: minItem.value,
            median: maxRaw.results[idx]?.value,
          }));

          transformed = transformMinMaxToChartJS(merged);
          break;
        }
        case 'range':
        case 'stddev': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [rangeRes, stddevRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/range?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/stddev?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
          ]);

          if (!rangeRes.ok || !stddevRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const rangeRaw = await rangeRes.json();
          const stddevRaw = await stddevRes.json();

          const merged = rangeRaw.results.map((rangeItem, idx) => ({
            label: rangeItem.label,
            mean: rangeItem.value,
            median: stddevRaw.results[idx]?.value,
          }));

          transformed = transformRangeStdDevToChartJS(merged);
          break;
        }
        case 'sum':
        case 'count': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [sumRes, countRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/sum?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/count?dataset_id=${datasetId}`, { credentials: "include" }),
          ]);

          if (!sumRes.ok || !countRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const sumRaw = await sumRes.json();
          const countRaw = await countRes.json();

          const merged = sumRaw.results.map((sumItem, idx) => ({
            label: sumItem.label,
            mean: sumItem.value,
            median: countRaw.results[idx]?.value,
          }));

          transformed = transformSumCountToChartJS(merged);
          break;
        }
        case 'stddev':
        case 'variance': {
          const baseUrl = "http://localhost:8080/analytics";
          const datasetId = activeDataset?.id;

          const [stdDevRes, varianceRes] = await Promise.all([
            fetch(`${baseUrl}/descriptives/stddev?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
            fetch(`${baseUrl}/descriptives/variance?dataset_id=${datasetId}&column=${defaultColumn}`, { credentials: "include" }),
          ]);

          if (!stdDevRes.ok || !varianceRes.ok) {
            throw new Error("Failed to fetch mean or median");
          }

          const stdDevRaw = await stdDevRes.json();
          const varianceRaw = await varianceRes.json();

          const merged = stdDevRaw.results.map((stdDevItem, idx) => ({
            label: stdDevItem.label,
            mean: stdDevItem.value,
            median: varianceRaw.results[idx]?.value,
          }));

          transformed = transformStdDevVarianceToChartJS(merged);
          break;
        }
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
          break;
        default:
          console.error(`No transform defined for method: ${method}`);
          throw new Error(`Unsupported analysis method: ${method}`);
      }
      const normalizedMethod = normalizeMethodKey(method);

      // Update chart state
      setLabels(transformed.labels);
      setDatasets(transformed.datasets);
      setTitle(transformed.title);
      setMethod(method);
    } catch (err) {
      if (err.message === "Failed to fetch" || err.message === "NetworkError") {
        try {
          await fetch("http://localhost:8080/user/logout", {
            method: "POST",
            credentials: "include",
          });
        } catch (_) {
          sessionStorage.setItem("logoutOnReload", "true");
        }

        setErrorDisabled(true);
        setUser(null);
        localStorage.removeItem("user");
        setTimeout(() => {
        alert("Network error, you have been logged out.");
      }, 50)
      }
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
            disabled={!activeDataset || errorDisabled || user === null}
            options={[
              {
                text: "Grouped Data",
                children: [
                  /*{ text: "Grouped By", onClick: () => handleSelect("aggregation", "grouped-by") },*/
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
            label="Correlation"
            disabled={!activeDataset || errorDisabled || user === null}
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
            disabled={!activeDataset || errorDisabled || user === null}
            options={[
              { text: "Mean/Median", onClick: () => handleSelect("descriptives", "mean") },
              { text: "Mode", onClick: () => handleSelect("descriptives", "mode") },
              { text: "Spread (Std Dev/Range)", onClick: () => handleSelect("descriptives", "stddev") },
              { text: "Variance", onClick: () => handleSelect("descriptives", "variance") },
              { text: "Min/Max", onClick: () => handleSelect("descriptives", "min") },
              { text: "Sum/Count", onClick: () => handleSelect("descriptives", "sum") },
            ]}
          />

          <DropdownMenu
            label="Distribution"
            disabled={!activeDataset || errorDisabled || user === null}
            options={[
              { text: "Histogram", onClick: () => handleSelect("distribution", "histogram") },
              { text: "KDE", onClick: () => handleSelect("distribution", "kde") },
            ]}
          />

          <DropdownMenu
            label="Outliers"
            disabled={!activeDataset || errorDisabled || user === null}
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
                onSuccess={async () => {
                  await fetchUserProfile();
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
