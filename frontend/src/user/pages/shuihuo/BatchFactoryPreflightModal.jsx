import { CheckCircleFilled, StopOutlined, WarningOutlined } from '@ant-design/icons';
import { Alert, Button, Modal, Space, Tag, Typography } from 'antd';

const { Text, Paragraph } = Typography;

const FIX_TARGET_LABELS = {
  batch_settings: '统一配置',
  book_settings: '单书设置',
  preset: '自动化预设',
  admin_models: '管理员 → 水货生产模型',
  publish_settings: '发布设置'
};

const TASK_STATE_LABELS = {
  running: '运行中',
  paused: '已暂停',
  scheduled: '已定时',
  needs_attention: '需要处理'
};

function fixLabel(issue) {
  return issue?.fixLabel || FIX_TARGET_LABELS[issue?.fixTarget] || '';
}

function IssueLine({ issue, bookId, onFix }) {
  return (
    <div className="batch-factory-preflight-issue">
      <Text type="danger">{issue.message}</Text>
      {issue.fixTarget ? (
        <Button size="small" type="link" onClick={() => onFix?.(issue.fixTarget, bookId)}>
          去修复：{fixLabel(issue)}
        </Button>
      ) : null}
    </div>
  );
}

function BookRow({ book, onFix }) {
  const title = book.title || book.bookId;
  return (
    <div className={`batch-factory-preflight-book${book.passed ? ' is-passed' : ' is-blocked'}`}>
      <div className="batch-factory-preflight-book-head">
        {book.passed ? <CheckCircleFilled style={{ color: 'var(--text-success)' }} /> : <StopOutlined style={{ color: 'var(--text-danger)' }} />}
        <Text strong ellipsis={{ tooltip: title }} style={{ marginLeft: 6 }}>{title}</Text>
        {book.passed ? <Tag color="success" style={{ marginLeft: 8 }}>通过</Tag> : <Tag color="error" style={{ marginLeft: 8 }}>阻断</Tag>}
        {book.overrides?.length ? <Tag color="warning" style={{ marginLeft: 4 }}>有单书覆盖</Tag> : null}
      </div>
      {book.overrides?.length ? (
        <Paragraph type="warning" style={{ margin: '4px 0 0' }}>
          这本书有单书级覆盖（{book.overrides.map(item => item.label || item.key).join('、')}），改统一配置对这本无效，请到「单书设置」修改。
        </Paragraph>
      ) : null}
      {book.blocking?.map((issue, index) => (
        <IssueLine key={`block-${issue.code}-${index}`} issue={issue} bookId={book.bookId} onFix={onFix} />
      ))}
      {book.warnings?.map((issue, index) => (
        <div className="batch-factory-preflight-issue is-warning" key={`warn-${issue.code}-${index}`}>
          <WarningOutlined style={{ color: 'var(--text-warning)', marginRight: 6 }} />
          <Text type="warning">{issue.message}</Text>
          {issue.fixTarget ? (
            <Button size="small" type="link" onClick={() => onFix?.(issue.fixTarget, book.bookId)}>
              去检查：{fixLabel(issue)}
            </Button>
          ) : null}
        </div>
      ))}
    </div>
  );
}

export function BatchFactoryPreflightModal({
  open = false,
  loading = false,
  result = null,
  error = '',
  submitting = false,
  onClose,
  onRun,
  onFix
}) {
  const books = Array.isArray(result?.books) ? result.books : [];
  const blockedBooks = books.filter(book => !book.passed);
  const passedBooks = books.filter(book => book.passed);
  const batchIssues = Array.isArray(result?.batchIssues) ? result.batchIssues : [];
  const skippedBooks = books.filter(book => book.skipped);
  const pendingBooks = passedBooks.filter(book => !book.skipped);
  const skippedCount = result?.summary?.skipped ?? skippedBooks.length;
  const passedCount = result?.summary?.passed ?? pendingBooks.length;
  const blockedCount = result?.summary?.blocked ?? blockedBooks.length;
  // 服务端只读 peek 会在预检返回体里带 taskState；有活动任务时打开面板就
  // 禁用提交，避免用户点完才收到 409。
  const taskState = result?.taskState || '';
  const taskActive = Boolean(taskState);
  const taskStateLabel = TASK_STATE_LABELS[taskState] || taskState;
  const runningSince = result?.runningSince || '';
  const taskMessage = runningSince
    ? `当前批次已有任务在进行（${taskStateLabel}，开始时间：${runningSince}）`
    : `当前批次已有任务在进行（${taskStateLabel}）`;
  const canRun = !batchIssues.length && pendingBooks.length > 0 && !taskActive;
  const allBlocked = books.length > 0 && passedBooks.length === 0;
  // 全阻断时不允许提交：面板只给「先修这几项」清单和关闭。
  const fixList = [...batchIssues, ...blockedBooks.flatMap(book => book.blocking || [])];

  const runLabel = taskActive
    ? '当前批次任务进行中，请先停止后再执行'
    : batchIssues.length
      ? '先修复整批配置问题'
      : pendingBooks.length === 0 && skippedCount > 0
        ? `所选小说均已完成，无需执行（${skippedCount} 本会跳过）`
        : blockedCount > 0
          ? `只跑通过的 ${passedCount} 本（${blockedCount} 本不会提交）`
          : `立即执行（${passedCount} 本${skippedCount > 0 ? `，跳过已完成 ${skippedCount} 本` : ''}）`;

  return (
    <Modal
      title="执行前质检"
      open={open}
      onCancel={onClose}
      width={720}
      destroyOnClose
      maskClosable={!submitting}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={submitting}>
          全部取消，先修配置
        </Button>,
        allBlocked && !taskActive ? null : (
          <Button
            key="run"
            type="primary"
            danger={blockedCount > 0}
            loading={submitting}
            disabled={loading || !canRun}
            onClick={() => onRun?.(passedBooks.map(book => book.bookId))}
          >
            {runLabel}
          </Button>
        )
      ]}
    >
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        {taskActive ? (
          <Alert
            type="warning"
            showIcon
            message={taskMessage}
            description="提交按钮已禁用。请先到自动生产面板停止当前任务，等状态变为已停止后再回来重试；任务运行期间重复提交会返回 409。"
          />
        ) : null}
        <Alert
          type={loading ? 'info' : error ? 'error' : allBlocked ? 'error' : blockedCount > 0 ? 'warning' : 'success'}
          showIcon
          message={loading ? '正在预检配置…' : error ? `预检失败：${error}` : `质检完成：通过 ${passedCount} 本 / 阻断 ${blockedCount} 本`}
          description="预检只读取配置，不会创建任务、不会修改批次或书籍设置。"
        />
        {batchIssues.length ? (
          <div>
            <Text strong>整批配置问题</Text>
            {batchIssues.map((issue, index) => (
              <IssueLine key={`batch-${issue.code}-${index}`} issue={issue} onFix={onFix} />
            ))}
          </div>
        ) : null}
        {!loading && !error && allBlocked ? (
          <div>
            <Text strong>先修这几项：</Text>
            {fixList.map((issue, index) => (
              <IssueLine key={`fix-${issue.code}-${index}`} issue={issue} onFix={onFix} />
            ))}
          </div>
        ) : null}
        {blockedCount > 0 && !allBlocked ? (
          <Alert
            type="warning"
            showIcon
            message={`这 ${blockedCount} 本不会提交：${blockedBooks.map(book => book.title || book.bookId).join('、')}`}
          />
        ) : null}
        {books.length ? (
          <div className="batch-factory-preflight-books">
            {books.map(book => (
              <BookRow key={book.bookId} book={book} onFix={onFix} />
            ))}
          </div>
        ) : null}
      </Space>
    </Modal>
  );
}

export default BatchFactoryPreflightModal;
