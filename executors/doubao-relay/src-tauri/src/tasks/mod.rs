//! 任务域:任务模型(`model`)、请求校验(`validate`)、调度器(`scheduler`)、运行器(`runner`);
//! 详情体(`detail`,GET /tasks/{id} 与 Webhook 共用);Webhook 回调(`webhook`)。

pub mod detail;
pub mod model;
pub mod runner;
pub mod scheduler;
pub mod validate;
pub mod webhook;

pub use model::{
    new_task_id, prompt_hash, ConfirmationMode, GenerateRequest, ResultFile, StageEntry, Task,
    TaskError, TaskStage, TaskStatus, TaskSummary, TaskType, WatermarkSource,
};
pub use runner::{DoubaoBackend, GenerationBackend};
pub use scheduler::{Scheduler, SchedulerTiming, TaskEvent, TaskHandle};
pub use validate::{
    ratios_for, validate_request, IMAGE_MODELS, IMAGE_RATIOS, RATIOS, VIDEO_MODELS, VIDEO_RATIOS,
};
