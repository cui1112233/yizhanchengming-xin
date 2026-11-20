/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801d8140 */

longlong FUN_1801d8140(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong lVar3;
  longlong *plVar4;
  char cVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong lVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  undefined8 uVar13;
  longlong local_res18 [2];

  plVar4 = DAT_1803bc8b8;
  plVar1 = (longlong *)*param_3;
  local_res18[0] = DAT_180391dd8;
  if (DAT_1803bc8b8 == (longlong *)0x0) {
LAB_1801d819b:
    DAT_1803bc8b8 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc860,DAT_1803ca5e8,0x10);
  }
  else if ((1 < *DAT_1803bc8b8) || (DAT_1803bc8b8[2] != 0)) {
    if ((DAT_1803bc8b8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bc8b8 && (*DAT_1803bc8b8 = *DAT_1803bc8b8 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1801d819b;
  }
  plVar4 = DAT_1803bc8b8;
  FUN_180001d00(param_1,DAT_1803bc8b8);
/* novels.prompts.inkos.short_fiction_prompts#72="根据审稿意见，继续写第二版完整正文。" */
  piVar2 = DAT_1803bcbb0;
  plVar6 = (longlong *)FUN_180372160(param_1,0x12);
  if (*piVar2 + 1 != 0) {
    *piVar2 = *piVar2 + 1;
  }
  *(int **)plVar6[3] = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#73="这是同一篇的第二轮写作：保留上一版能打的地方，修掉会让读者出戏或不想读的问题。" */
  piVar2 = DAT_1803bcbb8;
/* novels.prompts.inkos.short_fiction_prompts#73="这是同一篇的第二轮写作：保留上一版能打的地方，修掉会让读者出戏或不想读的问题。" */
  if (*DAT_1803bcbb8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#73="这是同一篇的第二轮写作：保留上一版能打的地方，修掉会让读者出戏或不想读的问题。" */
    *DAT_1803bcbb8 = *DAT_1803bcbb8 + 1;
  }
  *(int **)(plVar6[3] + 8) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#74="不要只列修改建议，不要只改几章片段，输出完整正文。" */
  piVar2 = DAT_1803bcbc0;
/* novels.prompts.inkos.short_fiction_prompts#74="不要只列修改建议，不要只改几章片段，输出完整正文。" */
  if (*DAT_1803bcbc0 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#74="不要只列修改建议，不要只改几章片段，输出完整正文。" */
    *DAT_1803bcbc0 = *DAT_1803bcbc0 + 1;
  }
  *(int **)(plVar6[3] + 0x10) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
  piVar2 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
  if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    *DAT_1803bc980 = *DAT_1803bc980 + 1;
  }
  *(int **)(plVar6[3] + 0x18) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#75="## 审稿意见" */
  piVar2 = DAT_1803bcbc8;
/* novels.prompts.inkos.short_fiction_prompts#75="## 审稿意见" */
  if (*DAT_1803bcbc8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#75="## 审稿意见" */
    *DAT_1803bcbc8 = *DAT_1803bcbc8 + 1;
  }
  *(int **)(plVar6[3] + 0x20) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#30="review" */
  plVar7 = (longlong *)FUN_18037f0c0(piVar2,plVar1,DAT_1803bca60);
  if (plVar7 == (longlong *)0x0) {
    lVar9 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1801d82f7:
    uVar13 = 0x11b;
    plVar7 = (longlong *)0x0;
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, plVar7 = (longlong *)0x0, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
  }
  else {
    plVar8 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) {
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801d82f7;
    }
    *(undefined4 *)(plVar4 + 5) = 0x11b;
    lVar9 = FUN_180376210(param_1,plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (lVar9 == 0) {
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801d82f7;
    }
    *(longlong *)(plVar6[3] + 0x28) = lVar9;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    piVar2 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar6[3] + 0x30) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#76="## 第二轮重点" */
    piVar2 = DAT_1803bcbd0;
/* novels.prompts.inkos.short_fiction_prompts#76="## 第二轮重点" */
    if (*DAT_1803bcbd0 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#76="## 第二轮重点" */
      *DAT_1803bcbd0 = *DAT_1803bcbd0 + 1;
    }
    *(int **)(plVar6[3] + 0x38) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#77="- 修时间线、逻辑、人物关系、证据权限、身体状态等会让读者出戏的问题。" */
    piVar2 = DAT_1803bcbd8;
/* novels.prompts.inkos.short_fiction_prompts#77="- 修时间线、逻辑、人物关系、证据权限、身体状态等会让读者出戏的问题。" */
    if (*DAT_1803bcbd8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#77="- 修时间线、逻辑、人物关系、证据权限、身体状态等会让读者出戏的问题。" */
      *DAT_1803bcbd8 = *DAT_1803bcbd8 + 1;
    }
    *(int **)(plVar6[3] + 0x40) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#78="- 补后半段有效场面，不要用结果摘要收尾。" */
    piVar2 = DAT_1803bcbe0;
/* novels.prompts.inkos.short_fiction_prompts#78="- 补后半段有效场面，不要用结果摘要收尾。" */
    if (*DAT_1803bcbe0 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#78="- 补后半段有效场面，不要用结果摘要收尾。" */
      *DAT_1803bcbe0 = *DAT_1803bcbe0 + 1;
    }
    *(int **)(plVar6[3] + 0x48) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#79="- 保持标题、开篇、章节标题和正文主标题一致，但标题可以基于正文重新压得更有平台点击感。" */
    piVar2 = DAT_1803bcbe8;
/* novels.prompts.inkos.short_fiction_prompts#79="- 保持标题、开篇、章节标题和正文主标题一致，但标题可以基于正文重新压得更有平台点击感。" */
    if (*DAT_1803bcbe8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#79="- 保持标题、开篇、章节标题和正文主标题一致，但标题可以基于正文重新压得更有平台点击感。" */
      *DAT_1803bcbe8 = *DAT_1803bcbe8 + 1;
    }
    *(int **)(plVar6[3] + 0x50) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#80="- 字数只做校准：偏短补有效场面，偏长删解释和重复反应。" */
    piVar2 = DAT_1803bcbf0;
/* novels.prompts.inkos.short_fiction_prompts#80="- 字数只做校准：偏短补有效场面，偏长删解释和重复反应。" */
    if (*DAT_1803bcbf0 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#80="- 字数只做校准：偏短补有效场面，偏长删解释和重复反应。" */
      *DAT_1803bcbf0 = *DAT_1803bcbf0 + 1;
    }
    *(int **)(plVar6[3] + 0x58) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    piVar2 = DAT_1803bc980;
/* novels.prompts.inkos.short_fiction_prompts#2="" */
    if (*DAT_1803bc980 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#2="" */
      *DAT_1803bc980 = *DAT_1803bc980 + 1;
    }
    *(int **)(plVar6[3] + 0x60) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
    piVar2 = DAT_1803bca68;
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
    if (*DAT_1803bca68 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#31="## 输出格式" */
      *DAT_1803bca68 = *DAT_1803bca68 + 1;
    }
    *(int **)(plVar6[3] + 0x68) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#49="=== SHORT_FICTION_TITLE ===" */
    piVar2 = DAT_1803bcaf8;
/* novels.prompts.inkos.short_fiction_prompts#49="=== SHORT_FICTION_TITLE ===" */
    if (*DAT_1803bcaf8 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#49="=== SHORT_FICTION_TITLE ===" */
      *DAT_1803bcaf8 = *DAT_1803bcaf8 + 1;
    }
    *(int **)(plVar6[3] + 0x70) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#50="短篇标题，只写纯文本平台标题" */
    piVar2 = DAT_1803bcb00;
/* novels.prompts.inkos.short_fiction_prompts#50="短篇标题，只写纯文本平台标题" */
    if (*DAT_1803bcb00 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#50="短篇标题，只写纯文本平台标题" */
      *DAT_1803bcb00 = *DAT_1803bcb00 + 1;
    }
    *(int **)(plVar6[3] + 0x78) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#51="=== SHORT_FICTION_OPENING_HOOK ===" */
    piVar2 = DAT_1803bcb08;
/* novels.prompts.inkos.short_fiction_prompts#51="=== SHORT_FICTION_OPENING_HOOK ===" */
    if (*DAT_1803bcb08 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#51="=== SHORT_FICTION_OPENING_HOOK ===" */
      *DAT_1803bcb08 = *DAT_1803bcb08 + 1;
    }
    *(int **)(plVar6[3] + 0x80) = piVar2;
/* novels.prompts.inkos.short_fiction_prompts#52="可选正文前小钩子，约 200 字；如果不需要独立引子，也要写第 1 章第一屏的入局小场面" */
    piVar2 = DAT_1803bcb10;
/* novels.prompts.inkos.short_fiction_prompts#52="可选正文前小钩子，约 200 字；如果不需要独立引子，也要写第 1 章第一屏的入局小场面" */
    if (*DAT_1803bcb10 + 1 != 0) {
/* novels.prompts.inkos.short_fiction_prompts#52="可选正文前小钩子，约 200 字；如果不需要独立引子，也要写第 1 章第一屏的入局小场面" */
      *DAT_1803bcb10 = *DAT_1803bcb10 + 1;
    }
    *(int **)(plVar6[3] + 0x88) = piVar2;
    lVar9 = FUN_1801d3910(param_1);
    plVar7 = plVar6;
    if (lVar9 == 0) {
/* novels.prompts.inkos.short_fiction_prompts#53="_chapter_output_blocks" */
      FUN_180373340(param_1,local_res18,DAT_1803bcb18);
      uVar13 = 0x129;
      lVar9 = local_res18[0];
    }
    else {
      plVar8 = (longlong *)FUN_18037f0c0();
      uVar13 = 0x129;
      if (plVar8 != (longlong *)0x0) {
        *(undefined4 *)(plVar4 + 5) = 0x129;
        plVar10 = (longlong *)FUN_180376770(param_1,lVar9,plVar8);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar10 != (longlong *)0x0) {
          cVar5 = FUN_180372360(param_1,plVar6,plVar10);
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
          if (cVar5 != '\0') {
/* novels.prompts.inkos.short_fiction_prompts#14="\n" */
            lVar9 = PyUnicode_Join(DAT_1803bc9e0,plVar6);
            if (lVar9 != 0) {
              FUN_180001d80(param_1);
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              if ((int)*plVar1 < 0) {
                return lVar9;
              }
              *plVar1 = *plVar1 + -1;
              if (*plVar1 != 0) {
                return lVar9;
              }
              (**(code **)(plVar1[1] + 0x30))(plVar1);
              return lVar9;
            }
            uVar13 = 0x12a;
          }
        }
      }
      lVar9 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
  }
  piVar2 = *(int **)(lVar9 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    plVar6 = *(longlong **)(lVar9 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
LAB_1801d8589:
      pcVar11 = pcVar12;
      if ((pcVar12 != (code *)0x0) && (*(int *)pcVar12 + 1 != 0)) {
        *(int *)pcVar12 = *(int *)pcVar12 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar4) goto LAB_1801d85d2;
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    *(int **)(pcVar11 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = *(longlong **)(lVar9 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 != _Py_NoneStruct_exref) goto LAB_1801d8589;
    pcVar12 = (code *)0x0;
  }
  *(code **)(lVar9 + 0x28) = pcVar12;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_1801d85d2:
  FUN_18036d0a0(plVar4,&DAT_180391be4,plVar1,plVar7);
  plVar6 = DAT_1803bc8b8;
  if (plVar4 == DAT_1803bc8b8) {
    if ((-1 < (int)*DAT_1803bc8b8) && (*DAT_1803bc8b8 = *DAT_1803bc8b8 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803bc8b8 = (longlong *)0x0;
  }
  lVar3 = **(longlong **)(param_1 + 0x38);
  plVar4 = *(longlong **)(lVar3 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar3 + 8);
  *(undefined4 *)(plVar4 + 8) = 0xffffffff;
  plVar6 = (longlong *)plVar4[2];
  if (((plVar6 != (longlong *)0x0) && (plVar4[2] = 0, -1 < (int)*plVar6)) &&
     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  plVar4[10] = 0;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar9;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
