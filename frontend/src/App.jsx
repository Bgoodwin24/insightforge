import React, { useState } from "react";
import Button from "./components/Button";
import Modal from "./components/Modal";
import AuthForm from "./components/AuthForm";
import UploadForm from "./components/UploadForm";
import AccountMenu from "./components/AccountMenu";
import * as styles from "./App.module.css";
import { transformStandardizeColumnForChart, transformBoxPlotDataToChartJS, transformCountToChartJS, transformDroppedRowsToChartJS, transformFilledMissingToChartJS, transformFilteredSortedDataToChartJS, transformGroupedDataToChartJS, transformHistogramToChartJS, transformIQROutliersToChartJS, transformKDEToChartJS, transformLogTransformedToChartJS, transformMaxToChartJS, transformMinToChartJS, transformModeToChartJS, transformNormalizedColumnToChartJS, transformPivotDataToChartJS, transformRangeToChartJS, transformSumToChartJS, transformVarianceToChartJS, transformZScoreOutliersToChartJS } from "./helpers/chartDataTransform";
import DatasetChart from "./components/charts/DatasetChart";
import { chartTypeByMethod } from "./components/charts/chartMethodMap";

export default function App() {
  const [modalType, setModalType] = useState(null);
  const [user, setUser] = useState(null); // Track the logged-in user
  const [method, setMethod] = useState(null);
  const [labels, setLabels] = useState([]);
  const [datasets, setDatasets] = useState([]);
  const [title, setTitle] = useState("");


  const closeModal = () => setModalType(null);

  const handleLogout = async () => {
    await fetch("http://localhost:8080/user/logout", {
      method: "POST",
      credentials: "include",
    });
    setUser(null);
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
        case 'KDE':
          transformed = transformKDEToChartJS(raw);
          break;
        case 'filterSort':
          transformed = transformFilteredSortedDataToChartJS(raw);
          break;
        case 'ZScore':
          transformed = transformZScoreOutliersToChartJS(raw);
          break;
        case 'IQR':
          transformed = transformIQROutliersToChartJS(raw);
          break;
        case 'boxPlot':
          transformed = transformBoxPlotDataToChartJS(raw);
          break;
        case 'groupedData':
          transformed = transformGroupedDataToChartJS(raw);
          break;
        case 'pivotData':
          transformed = transformPivotDataToChartJS(raw);
          break;
        case 'droppedRow':
          transformed = transformDroppedRowsToChartJS(raw);
          break;
        case 'fillMissingWith':
          transformed = transformFilledMissingToChartJS(raw);
          break;
        case 'logTransform':
          transformed = transformLogTransformedToChartJS(raw);
          break;
        case 'normalizeColumn':
          transformed = transformNormalizedColumnToChartJS(raw);
          break;
        case 'standardizeColumn':
          transformed = transformStandardizeColumnForChart(raw);
          break;
        case 'correlation-pearson':
          transformed = transformPearsonForChartJS(raw);
          break;
        case 'correlation-spearman':
          transformed = transformSpearmanForChartJS(raw);
          break;
        case 'correlationMatrix':
          transformed = transformCorrelationMatrixForChart(raw);
          break;
        default:
          throw new Error(`No transform defined for method: ${method}`);
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
            <Button text="Login / Register" onClick={() => setModalType("auth")} />
          ) : (
            <div className={styles.accountMenu}>
              <AccountMenu
                username={user.username}
                email={user.email}
                onLogout={handleLogout}
              />
            </div>
          )}
          <Button text="Upload Dataset" onClick={() => setModalType("upload")} />
        </div>
      </header>

      <div className={styles.divider}></div>

      {/* Content with Sidebar */}
      <div className={styles.contentWrapper}>
        <aside className={styles.sidebar}>
          <h3>Forge Tools</h3>
        </aside>

        <main className={styles.mainContent}>
          <div className={styles.mainText}>Explore your datasets here.</div>

           <div className={styles.chartContainer}>
              <DatasetChart
                chartType={chartTypeByMethod[method]}
                labels={labels}
                datasets={datasets}
                title={title}
              />
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
