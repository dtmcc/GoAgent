1. Agent 是把模型Model、工具Tools、记忆Memory、知识Knowledge，RAG、计划Planing、循环Loop 组织起来的**应用结构**，跟”换一个模型"不是一回事。 
2. ReAct = 思考Thought-Model、行动Action-Tools、观察Observation-Memory,Knowledge、汇总Final
3. Planing画路线图 ReAct执行每一步
4. Plan-and-Execute
  - Planner：一次性生成完整计划（通常是一次模型调用，Prompt 专注拆解） 
  - Executor：按计划逐步执行（可以是另一个 Agent，也可以是一组工具调用）
  - Replanner：执行过程中出问题了，回头调整后续步骤
5. 高级Planing
  - Tree of Thoughts（思维树）：在每一步生成多个候选思路，给每个思路打分，只
    保留得分最高的几个继续往下想。适合“答案质量比成本重要"的场景，比如系统架构
    设计、关键决策分析。
  - Reflection（反思 /自我审视）：让 Agent 生成方案后，再用一个"Critic"角色挑毛
    病，反复迭代直到通过。Generator + Critic 是常见的双 Agent 配合方式。
6. Memory
  - Session 短期记忆，当前会话  短期上下文 -- 最近这一轮聊了什么
  - Profile 用户画像，长期记忆  长期、跨会话 -- 用户身份，水平，目标，阶段等
  - Summary 摘要记忆，最近摘要  压缩出来的状态 -- 最近卡点，给过的建议，下一步等 
  - 请求流程
    1. 从DB 拉 Profile 
    2. 从 DB 拉最近一次 Summary
    3. 从内存 / 短期表里拉当前 Session 的最近几轮
    4. 拼成 Prompt -> 调模型
    5. 模型回答完，看是否需要更新 Summary 或 Profile
    6. 写回 DB