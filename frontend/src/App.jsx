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

export default function App() {
  const [modalType, setModalType] = useState(null);
  const [user, setUser] = useState(null); // Track the logged-in user
  const [method, setMethod] = useState(null);
  const [labels, setLabels] = useState([]);
  const [datasets, setDatasets] = useState([]);
  const [title, setTitle] = useState("");
  const [datasetId, setDatasetId] = useState(null);


  const closeModal = () => setModalType(null);

    useEffect(() => {
      // On mount, try fetching the logged-in user profile
      const fetchUserProfile = async () => {
        try {
          const res = await fetch("/user/profile", {
            credentials: "include", // to send cookies
          });
          if (res.ok) {
            const userData = await res.json();
            setUser(userData.user);
          } else {
            // not logged in or no session, do nothing or setUser(null)
            setUser(null);
          }
        } catch (error) {
          console.error("Failed to fetch user profile:", error);
          setUser(null);
        }
      };

      fetchUserProfile();
    }, []);

  const handleLogout = async () => {
    try {
      await fetch("/user/logout", {
        method: "POST",
        credentials: "include", // important to send cookies
      });
      setUser(null);
    } catch (error) {
      console.error("Logout failed", error);
    }
  };


  const handleSelect = async (group, method) => {
    const url = `http://localhost:3000/analytics/${group}/${method}?dataset_id=${datasetId}`;

    try {
      const res = await fetch(url);
      const raw = await res.json();

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
          transformed = transformZScoreOutliersToChartJS(raw);
          break;
        case 'iqr-outliers':
          transformed = transformIQROutliersToChartJS(raw);
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
          transformed = transformGroupedDataToChartJS(raw);
          break;
        case 'pivot-sum':
        case 'pivot-mean':
        case 'pivot-count':
        case 'pivot-min':
        case 'pivot-max':
        case 'pivot-median':
        case 'pivot-stddev':
          transformed = transformPivotDataToChartJS(raw);
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
          transformed = transformPearsonForChartJS(raw);
          break;
        case 'spearman-correlation':
          transformed = transformSpearmanForChartJS(raw);
          break;
        case 'correlation-matrix':
          transformed = transformCorrelationMatrixForChart(raw);
          break;
        default:
          console.error(`No transform defined for method: ${method}`);
          throw new Error(`Unsupported analysis method: ${method}`);
      }

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

        <div className={styles.buttonContainer}>
          {!user ? (
            <Button text="Login / Register" onClick={() => setModalType("auth") } />
          ) : (
            <div className={styles.accountMenu}>
              <AccountMenu
                username={user.username}
                email={user.email}
                onLogout={handleLogout}
              />
            </div>
          )}
          <Button
            text="Upload Dataset"
            onClick={() => setModalType("upload")}
            disabled={!user}
            title={!user ? "Login to upload a dataset" : ""}
          />
        </div>
      </header>

      <div className={styles.divider}></div>

      {/* Content with Sidebar */}
      <div className={styles.contentWrapper}>
        <aside className={styles.sidebar}>
          <h3>Forge Tools</h3>

          <DropdownMenu
            label="Aggregation"
            disabled={datasetId === null}
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
              disabled={datasetId === null}
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
            disabled={datasetId === null}
            options={[
              { text: "Pearson", onClick: () => handleSelect("correlation", "pearson-correlation") },
              { text: "Spearman", onClick: () => handleSelect("correlation", "spearman-correlation") },
              { text: "Correlation Matrix", onClick: () => handleSelect("correlation", "correlation-matrix") },
            ]}
          />

          <DropdownMenu
            label="Descriptive Statistics"
            disabled={datasetId === null}
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
            disabled={datasetId === null}
            options={[
              { text: "Histogram", onClick: () => handleSelect("distribution", "histogram") },
              { text: "KDE", onClick: () => handleSelect("distribution", "kde") },
            ]}
          />

          <DropdownMenu
            label="Filter Sort"
            disabled={datasetId === null}
            options={[
              { text: "Filter Sort", onClick: () => handleSelect("filtersort", "filter-sort") },
            ]}
          />

          <DropdownMenu
            label="Outliers"
            disabled={datasetId === null}
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
                  chartType={chartTypeByMethod[method]}
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
            <Modal title="Upload Dataset" onClose={closeModal}>
              <UploadForm onSuccess={closeModal} />
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
