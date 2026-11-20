/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d8730 */

longlong FUN_1801d8730(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  undefined8 *puVar2;
  int *piVar3;
  longlong *plVar4;
  undefined8 uVar5;
  longlong *plVar6;
  longlong lVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  undefined8 uVar12;

  plVar4 = DAT_1803bc8d0;
  plVar1 = (longlong *)*param_3;
  if (DAT_1803bc8d0 == (longlong *)0x0) {
LAB_1801d877f:
    DAT_1803bc8d0 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc870,DAT_1803ca5e8,8);
  }
  else if ((1 < *DAT_1803bc8d0) || (DAT_1803bc8d0[2] != 0)) {
    if ((DAT_1803bc8d0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc8d0 && (*DAT_1803bc8d0 = *DAT_1803bc8d0 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1801d877f;
  }
  plVar4 = DAT_1803bc8d0;
  FUN_180001d00(param_1,DAT_1803bc8d0);
/* novels.prompts.inkos.short_fiction_prompts#14="\n" */
  uVar5 = DAT_1803bc9e0;
/* novels.prompts.inkos.short_fiction_prompts#0="## 创作方向" */
  piVar3 = DAT_1803bc970;
  plVar6 = (longlong *)FUN_180372160(param_1,0x12);
  if (*piVar3 + 1 != 0) {
    *piVar3 = *piVar3 + 1;
  }
  puVar2 = (undefined8 *)plVar6[3];
  *puVar2 = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#1="direction" */
  lVar7 = FUN_18037f0c0(puVar2,plVar1,DAT_1803bc978);
  if (lVar7 == 0) {
    uVar12 = 0x138;
LAB_1801d89f9:
    lVar7 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
  }
  else {
    *(longlong *)(plVar6[3] + 8) = lVar7;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    piVar3 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar6[3] + 0x10) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#47="## 故事方案" */
    piVar3 = DAT_1803bcae8;
/* novels.prompts.inkos.short_fiction_prompts#47="## 故事方案" */
    if (*DAT_1803bcae8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#47="## 故事方案" */
      *DAT_1803bcae8 = *DAT_1803bcae8 + 1;
    }
    *(int **)(plVar6[3] + 0x18) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#48="outline_markdown" */
    plVar8 = (longlong *)FUN_18037f0c0(piVar3,plVar1,DAT_1803bcaf0);
    if (plVar8 == (longlong *)0x0) {
      uVar12 = 0x13b;
      goto LAB_1801d89f9;
    }
    plVar9 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    uVar12 = 0x13b;
    if (plVar9 == (longlong *)0x0) goto LAB_1801d89f9;
    *(undefined4 *)(plVar4 + 5) = 0x13b;
    lVar7 = FUN_180376210(param_1,plVar9);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (lVar7 == 0) goto LAB_1801d89f9;
    *(longlong *)(plVar6[3] + 0x20) = lVar7;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    piVar3 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar6[3] + 0x28) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#81="## 最终正文" */
    piVar3 = DAT_1803bcbf8;
/* novels.prompts.inkos.short_fiction_prompts#81="## 最终正文" */
    if (*DAT_1803bcbf8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#81="## 最终正文" */
      *DAT_1803bcbf8 = *DAT_1803bcbf8 + 1;
    }
    *(int **)(plVar6[3] + 0x30) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#68="draft_markdown" */
    plVar8 = (longlong *)FUN_18037f0c0(piVar3,plVar1,DAT_1803bcb90);
    if (plVar8 == (longlong *)0x0) {
      uVar12 = 0x13e;
      goto LAB_1801d89f9;
    }
    plVar9 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    uVar12 = 0x13e;
    if (plVar9 == (longlong *)0x0) goto LAB_1801d89f9;
    *(undefined4 *)(plVar4 + 5) = 0x13e;
    lVar7 = FUN_180376210(param_1,plVar9);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (lVar7 == 0) goto LAB_1801d89f9;
    *(longlong *)(plVar6[3] + 0x38) = lVar7;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    piVar3 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar6[3] + 0x40) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
    piVar3 = DAT_1803bca68;
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
    if (*DAT_1803bca68 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
      *DAT_1803bca68 = *DAT_1803bca68 + 1;
    }
    *(int **)(plVar6[3] + 0x48) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#82="=== SHORT_FICTION_PACKAGE_TITLE ===" */
    piVar3 = DAT_1803bcc00;
/* novels.prompts.inkos.short_fiction_prompts#82="=== SHORT_FICTION_PACKAGE_TITLE ===" */
    if (*DAT_1803bcc00 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#82="=== SHORT_FICTION_PACKAGE_TITLE ===" */
      *DAT_1803bcc00 = *DAT_1803bcc00 + 1;
    }
    *(int **)(plVar6[3] + 0x50) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#83="draft_title" */
    lVar7 = FUN_18037f0c0(piVar3,plVar1,DAT_1803bcc08);
    if (lVar7 == 0) {
      uVar12 = 0x142;
      goto LAB_1801d89f9;
    }
    *(longlong *)(plVar6[3] + 0x58) = lVar7;
/* novels.prompts.inkos.short_fiction_prompts#84="=== SHORT_FICTION_INTRO ===" */
    piVar3 = DAT_1803bcc10;
/* novels.prompts.inkos.short_fiction_prompts#84="=== SHORT_FICTION_INTRO ===" */
    if (*DAT_1803bcc10 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#84="=== SHORT_FICTION_INTRO ===" */
      *DAT_1803bcc10 = *DAT_1803bcc10 + 1;
    }
    *(int **)(plVar6[3] + 0x60) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#85="100-180字平台简介，直接抓冲突、压迫和回报，不要剧透成流水账。" */
    piVar3 = DAT_1803bcc18;
/* novels.prompts.inkos.short_fiction_prompts#85="100-180字平台简介，直接抓冲突、压迫和回报，不要剧透成流水账。" */
    if (*DAT_1803bcc18 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#85="100-180字平台简介，直接抓冲突、压迫和回报，不要剧透成流水账。" */
      *DAT_1803bcc18 = *DAT_1803bcc18 + 1;
    }
    *(int **)(plVar6[3] + 0x68) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#86="=== SHORT_FICTION_SELLING_POINTS ===" */
    piVar3 = DAT_1803bcc20;
/* novels.prompts.inkos.short_fiction_prompts#86="=== SHORT_FICTION_SELLING_POINTS ===" */
    if (*DAT_1803bcc20 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#86="=== SHORT_FICTION_SELLING_POINTS ===" */
      *DAT_1803bcc20 = *DAT_1803bcc20 + 1;
    }
    *(int **)(plVar6[3] + 0x70) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#87="- 3到6条卖点，每条一行" */
    piVar3 = DAT_1803bcc28;
/* novels.prompts.inkos.short_fiction_prompts#87="- 3到6条卖点，每条一行" */
    if (*DAT_1803bcc28 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#87="- 3到6条卖点，每条一行" */
      *DAT_1803bcc28 = *DAT_1803bcc28 + 1;
    }
    *(int **)(plVar6[3] + 0x78) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#88="=== SHORT_FICTION_COVER_PROMPT ===" */
    piVar3 = DAT_1803bcc30;
/* novels.prompts.inkos.short_fiction_prompts#88="=== SHORT_FICTION_COVER_PROMPT ===" */
    if (*DAT_1803bcc30 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#88="=== SHORT_FICTION_COVER_PROMPT ===" */
      *DAT_1803bcc30 = *DAT_1803bcc30 + 1;
    }
    *(int **)(plVar6[3] + 0x80) = piVar3;
/* novels.prompts.inkos.short_fiction_prompts#89="中文封面生成提示词：3:4竖图，主标题区，人物情绪，道具，配色，字体风格，避免事项。" */
    piVar3 = DAT_1803bcc38;
/* novels.prompts.inkos.short_fiction_prompts#89="中文封面生成提示词：3:4竖图，主标题区，人物情绪，道具，配色，字体风格，避免事项。" */
    if (*DAT_1803bcc38 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#89="中文封面生成提示词：3:4竖图，主标题区，人物情绪，道具，配色，字体风格，避免事项。" */
      *DAT_1803bcc38 = *DAT_1803bcc38 + 1;
    }
    *(int **)(plVar6[3] + 0x88) = piVar3;
    lVar7 = PyUnicode_Join(uVar5,plVar6);
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    if (lVar7 != 0) {
      FUN_180001d80(param_1);
      if ((int)*plVar1 < 0) {
        return lVar7;
      }
      *plVar1 = *plVar1 + -1;
      if (*plVar1 != 0) {
        return lVar7;
      }
      (**(code **)(plVar1[1] + 0x30))(plVar1);
      return lVar7;
    }
    lVar7 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar12 = 0x136;
  }
  pcVar11 = (code *)0x0;
  piVar3 = *(int **)(lVar7 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar4,uVar12);
    plVar6 = *(longlong **)(lVar7 + 0x28);
    if (pcVar10 != _Py_NoneStruct_exref) {
LAB_1801d8b62:
      pcVar11 = pcVar10;
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1801d8bab;
    pcVar10 = (code *)FUN_18038e2a0(plVar4,uVar12);
    *(int **)(pcVar10 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar6 = *(longlong **)(lVar7 + 0x28);
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1801d8b62;
  }
  *(code **)(lVar7 + 0x28) = pcVar11;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1801d8bab:
  FUN_18036d0a0(plVar4,&DAT_180394940,plVar1);
  plVar6 = DAT_1803bc8d0;
  if (plVar4 == DAT_1803bc8d0) {
    if ((-1 < (int)*DAT_1803bc8d0) && (*DAT_1803bc8d0 = *DAT_1803bc8d0 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803bc8d0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar7;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
