# OpenClaw Skill 开发指南（高级）

## 📌 什么时候需要 Skill？

**需要 Skill 的场景：**
- ✅ 需要复杂的前置处理逻辑
- ✅ 需要调用多个 API 组合结果
- ✅ 需要访问数据库或外部系统
- ✅ 需要自定义认证和授权
- ✅ 需要缓存和性能优化

**不需要 Skill 的场景（用方案一即可）：**
- ❌ 简单的对话问答
- ❌ 直接的工具调用
- ❌ 标准的 AI 诊断

---

## 🔨 编写 Siriusec MCP Custom Skill

### 示例：创建一个增强的诊断 Skill

```python
# skills/enhanced_diagnosis.py
"""
OpenClaw Skill - 增强版服务器诊断
功能：
1. 自动收集系统信息
2. 调用多个诊断工具
3. 综合分析并给出建议
"""

from openclaw.skill import BaseSkill
from openclaw.decorators import skill
import requests
import json

class EnhancedDiagnosisSkill(BaseSkill):
    """增强诊断 Skill"""
    
    def __init__(self, config):
        super().__init__(config)
        self.mcp_server = config.get('mcp_server', 'http://localhost:7140')
        self.timeout = config.get('timeout', 120)
    
    @skill(name="comprehensive_diagnosis", 
           description="综合诊断服务器问题，自动调用多个工具")
    async def comprehensive_diagnosis(self, symptom: str, context: str = ""):
        """
        综合诊断
        
        Args:
            symptom: 症状描述
            context: 上下文信息
        """
        
        # 步骤 1: 收集系统信息
        system_info = await self.collect_system_info()
        
        # 步骤 2: 根据症状选择工具
        tools_to_call = self.select_tools(symptom)
        
        # 步骤 3: 并行调用多个工具
        results = await self.call_tools_parallel(tools_to_call, {
            'symptom': symptom,
            'context': context,
            'system_info': system_info
        })
        
        # 步骤 4: 综合分析
        analysis = await self.analyze_results(results)
        
        # 步骤 5: 生成建议
        recommendations = await self.generate_recommendations(analysis)
        
        return {
            'diagnosis': analysis,
            'recommendations': recommendations,
            'raw_data': results
        }
    
    async def collect_system_info(self):
        """收集系统基本信息"""
        try:
            response = requests.get(
                f'{self.mcp_server}/health',
                timeout=10
            )
            return response.json()
        except Exception as e:
            return {'error': str(e)}
    
    def select_tools(self, symptom: str) -> list:
        """根据症状选择要调用的工具"""
        
        symptom_lower = symptom.lower()
        
        tool_mapping = {
            'memory': ['memgraph', 'javamem', 'oomcheck'],
            'cpu': ['loadtask', 'delay'],
            'network': ['packetdrop', 'netjitter'],
            'io': ['iofsstat', 'iodiagnose'],
            'disk': ['disk_analysis'],
            'crash': ['analyze_vmcore', 'list_vmcores']
        }
        
        selected = []
        for keyword, tools in tool_mapping.items():
            if keyword in symptom_lower:
                selected.extend(tools)
        
        # 如果没有匹配，返回通用诊断工具
        if not selected:
            selected = ['smart_diagnose']
        
        return selected
    
    async def call_tools_parallel(self, tools: list, params: dict):
        """并行调用多个工具"""
        import asyncio
        
        async def call_single_tool(tool_name: str):
            try:
                request = {
                    'jsonrpc': '2.0',
                    'id': hash(tool_name),
                    'method': 'tools/call',
                    'params': {
                        'name': tool_name,
                        'arguments': params
                    }
                }
                
                response = requests.post(
                    f'{self.mcp_server}/mcp/unified',
                    json=request,
                    timeout=self.timeout
                )
                
                return {
                    'tool': tool_name,
                    'result': response.json(),
                    'status': 'success'
                }
            except Exception as e:
                return {
                    'tool': tool_name,
                    'error': str(e),
                    'status': 'failed'
                }
        
        # 并行执行
        tasks = [call_single_tool(tool) for tool in tools]
        results = await asyncio.gather(*tasks)
        
        return {r['tool']: r for r in results}
    
    async def analyze_results(self, results: dict):
        """分析工具返回结果"""
        # 这里可以使用 LLM 进行深度分析
        analysis_prompt = "请分析以下诊断结果，总结问题原因：\n\n"
        
        for tool_name, result in results.items():
            if result['status'] == 'success':
                analysis_prompt += f"{tool_name}: {json.dumps(result['result'])}\n"
        
        # 调用 LLM 分析
        llm_response = await self.call_llm(analysis_prompt)
        
        return llm_response
    
    async def generate_recommendations(self, analysis: str):
        """生成优化建议"""
        rec_prompt = f"基于以下分析，给出具体的优化建议：\n{analysis}"
        return await self.call_llm(rec_prompt)
    
    async def call_llm(self, prompt: str):
        """调用 OpenClaw LLM"""
        try:
            response = requests.post(
                f'{self.config["openclaw_url"]}/v1/chat/completions',
                json={
                    'model': self.config.get('model', 'qwen-7b'),
                    'messages': [{'role': 'user', 'content': prompt}],
                    'max_tokens': 2000
                },
                timeout=self.timeout
            )
            
            data = response.json()
            return data['choices'][0]['message']['content']
        except Exception as e:
            return f"LLM 调用失败：{e}"


# 注册 Skill
def register_skills(skill_registry):
    """注册所有 Skill"""
    skill_registry.register('enhanced_diagnosis', EnhancedDiagnosisSkill)
```

---

## 🔧 在 OpenClaw 中注册 Skill

### OpenClaw 配置文件

```yaml
# openclaw_config.yaml
skills:
  enhanced_diagnosis:
    module: skills.enhanced_diagnosis
    config:
      mcp_server: http://localhost:7140
      timeout: 120
      openclaw_url: http://localhost:8000
      model: qwen-7b
```

### 启动 OpenClaw（带 Skill）

```bash
openclaw start \
  --config openclaw_config.yaml \
  --skills-dir ./skills
```

---

## 🎯 使用方式对比

### 不使用 Skill（推荐）

```bash
# 简单配置即可
cp .env.openclaw .env
./start-webui.sh

# Web UI 直接使用
用户："帮我诊断 CPU 问题"
→ 自动调用 smart_diagnose
→ 返回结果
```

### 使用 Skill

```python
# 需要编写代码
from openclaw import Client

client = Client(config='openclaw_config.yaml')

# 调用自定义 Skill
result = client.skills.enhanced_diagnosis.comprehensive_diagnosis(
    symptom="CPU 使用率 100%",
    context="生产环境服务器"
)

print(result['diagnosis'])
print(result['recommendations'])
```

---

## 💡 建议

### 对于 90% 的场景：
✅ **使用方案一（直接调用 API）**
- 无需开发
- 开箱即用
- 维护成本低

### 仅在以下情况使用 Skill：
- 需要非常复杂的业务逻辑
- 需要集成多个系统
- 需要高度定制化

---

## 🚀 快速开始（方案一）

```bash
# 3 步完成集成
cd /Users/xiaoxi/Downloads/workspace/siriusec_mcp
cp .env.openclaw .env
# 编辑 .env 填入 OpenClaw 地址
./start-webui.sh
```

**就这么简单！不需要写任何 Skill！**
