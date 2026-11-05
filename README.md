- Agent 是把模型Model、工具Tools、记忆Memory、知识Knowledge，RAG、计划Planing、循环Loop 组织起来的**应用结构**，跟”换一个模型"不是一回事。 
- ReAct = 思考Thought-Model、行动Action-Tools、观察Observation-Memory,Knowledge、汇总Final
- Planing画路线图 ReAct执行每一步
- Plan-and-Execute
  - Planner：一次性生成完整计划（通常是一次模型调用，Prompt 专注拆解） 
  - Executor：按计划逐步执行（可以是另一个 Agent，也可以是一组工具调用）
  - Replanner：执行过程中出问题了，回头调整后续步骤
- 高级Planing
  - Tree of Thoughts（思维树）：在每一步生成多个候选思路，给每个思路打分，只
    保留得分最高的几个继续往下想。适合“答案质量比成本重要"的场景，比如系统架构
    设计、关键决策分析。
  - Reflection（反思 /自我审视）：让 Agent 生成方案后，再用一个"Critic"角色挑毛
    病，反复迭代直到通过。Generator + Critic 是常见的双 Agent 配合方式。