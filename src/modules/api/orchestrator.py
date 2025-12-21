#!/usr/bin/env python3
import uuid
from datetime import datetime, timezone
from typing import Optional
from fastapi import HTTPException

from api.worker_handler import WorkerHandler
from models.pipeline import Pipeline
from models.pipeline_run import PipelineRun, NodeRunState
from models.types import NodeStatus
from models.job_payload import JobPayload
from models.worker_result import WorkerResult
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

class Orchestrator:
    def __init__(self):
        self.logger = LoggingModule.get_logger()
        self.es_client = get_elasticsearch_connection()
        self.worker_handler = WorkerHandler()

    def start_pipeline(self, pipeline_id: str) -> PipelineRun:
        try:
            # 1. Fetch Pipeline
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            if not pipeline.trigger:
                raise HTTPException(status_code=400, detail="Pipeline has no trigger")

            # 2. Create PipelineRun
            run = PipelineRun(pipeline_id=pipeline.id)
            
            # Initialize states for all nodes (trigger + workers)
            run.node_states[str(pipeline.trigger.id)] = NodeRunState(status=NodeStatus.PENDING)
            if pipeline.worker:
                for worker in pipeline.worker:
                    run.node_states[str(worker.id)] = NodeRunState(status=NodeStatus.PENDING)

            # Save Run to ES
            self._save_run(run)

            # 3. Start Trigger
            self._start_node(run, pipeline, str(pipeline.trigger.id), None)

            return run
        except Exception as e:
            self.logger.error("Error starting pipeline %s: %s", pipeline_id, e)
            raise e

    def handle_node_completion(self, result: WorkerResult):
        try:
            run_id = str(result.run_id)
            node_id = str(result.node_id)

            # 1. Fetch Run
            run = self._get_run(run_id)
            if not run:
                self.logger.error("Run %s not found", run_id)
                return

            # 2. Update Node Status
            if node_id in run.node_states:
                run.node_states[node_id].status = result.status
                run.node_states[node_id].end_time = datetime.now(timezone.utc)
                if result.error:
                    run.node_states[node_id].error = result.error
            else:
                self.logger.warning("Node %s not found in run %s", node_id, run_id)

            self._save_run(run)

            # 3. Save Raw Results
            if result.raw_data:
                self._save_results(run, node_id, result.raw_data)

            # 4. If successful, trigger next nodes
            if result.status == NodeStatus.COMPLETED:
                # Fetch Pipeline to find edges
                pipeline_result = self.es_client.get(index="pipelines", id=str(run.pipeline_id))
                pipeline = Pipeline(**pipeline_result["_source"])

                # Find next nodes
                next_nodes = []
                current_node_uuid = uuid.UUID(node_id)
                
                if pipeline.edges:
                    for edge in pipeline.edges:
                        if edge.source == current_node_uuid:
                            next_nodes.append(str(edge.target))

                # Start next nodes
                for target_id in next_nodes:
                    self._start_node(run, pipeline, target_id, result.output_payload)

            # 5. Check if pipeline is finished
            self._check_pipeline_completion(run)

        except Exception as e:
            self.logger.error("Error handling node completion for run %s: %s", run_id, e)
            raise e

    def _save_results(self, run: PipelineRun, node_id: str, raw_data: dict):
        try:
            doc = {
                "run_id": str(run.id),
                "pipeline_id": str(run.pipeline_id),
                "node_id": node_id,
                "timestamp": datetime.now(timezone.utc),
                "data": raw_data
            }
            # Use a dedicated index for results, or one per pipeline?
            # Let's use a common index "pipeline_results" for now.
            self.es_client.index(index="pipeline_results", document=doc)
        except Exception as e:
            self.logger.error("Error saving results for node %s: %s", node_id, e)

    def _start_node(self, run: PipelineRun, pipeline: Pipeline, node_id: str, payload: Optional[JobPayload]):
        try:
            # Update state
            run.node_states[node_id].status = NodeStatus.RUNNING
            run.node_states[node_id].start_time = datetime.now(timezone.utc)
            self._save_run(run)

            # Find the node object
            node = None
            if pipeline.trigger and str(pipeline.trigger.id) == node_id:
                node = pipeline.trigger
                # Triggers usually don't take payload input in this context, 
                # but if it's a chained trigger/worker, it might.
                # For now, we assume triggers are self-starting or started by orchestrator without payload.
                self.worker_handler.start_trigger(str(pipeline.id), node, str(run.id))
            elif pipeline.worker:
                for w in pipeline.worker:
                    if str(w.id) == node_id:
                        node = w
                        break
                if node:
                    self.worker_handler.start_worker(str(pipeline.id), node, payload, str(run.id))
            
            if not node:
                self.logger.error("Node %s object not found in pipeline", node_id)
                run.node_states[node_id].status = NodeStatus.ERROR
                run.node_states[node_id].error = f"Node {node_id} object not found in pipeline"
                self._save_run(run)

        except Exception as e:
            self.logger.error("Error starting node %s: %s", node_id, e)
            run.node_states[node_id].status = NodeStatus.ERROR
            run.node_states[node_id].error = str(e)
            self._save_run(run)

    def _check_pipeline_completion(self, run: PipelineRun):
        # Simple check: if all nodes are COMPLETED or ERROR, pipeline is done.
        # Or if no nodes are RUNNING or PENDING.
        
        is_active = False
        for state in run.node_states.values():
            if state.status in [NodeStatus.RUNNING, NodeStatus.PENDING]:
                is_active = True
                break
        
        if not is_active:
            run.status = NodeStatus.COMPLETED
            run.end_time = datetime.now(timezone.utc)
            self._save_run(run)
            self.logger.info("Pipeline Run %s completed.", run.id)

    def _save_run(self, run: PipelineRun):
        self.es_client.index(index="pipeline_runs", id=str(run.id), document=run.model_dump(), refresh="wait_for")

    def _get_run(self, run_id: str) -> Optional[PipelineRun]:
        try:
            result = self.es_client.get(index="pipeline_runs", id=run_id)
            return PipelineRun(**result["_source"])
        except Exception:
            return None
