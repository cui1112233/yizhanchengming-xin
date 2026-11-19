/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801c8710 */

longlong FUN_1801c8710(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong *plVar6;
  int *piVar7;
  undefined8 *puVar8;
  longlong *plVar9;
  int *piVar10;
  int iVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  longlong *plVar15;
  longlong lVar16;
  longlong *plVar17;
  longlong *plVar18;
  longlong *plVar19;
  longlong *plVar20;
  undefined8 uVar21;
  longlong *plVar22;
  longlong *plVar24;
  longlong *plVar25;
  longlong *local_78;
  longlong *plVar23;
  
  plVar9 = DAT_1803bbd20;
  plVar1 = (longlong *)param_3[1];
  plVar2 = (longlong *)*param_3;
  plVar3 = (longlong *)param_3[3];
  plVar25 = (longlong *)0x0;
  plVar4 = (longlong *)param_3[4];
  plVar20 = (longlong *)0x0;
  plVar23 = (longlong *)0x0;
  plVar22 = (longlong *)0x0;
  plVar5 = (longlong *)param_3[2];
  plVar19 = (longlong *)0x0;
  plVar24 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  plVar6 = (longlong *)param_3[5];
  if (DAT_1803bbd20 == (longlong *)0x0) {
LAB_1801c87b9:
    DAT_1803bbd20 = (longlong *)FUN_18036cd10(param_1,DAT_1803bbd98,DAT_1803ca610,0x58);
  }
  else if ((1 < *DAT_1803bbd20) || (DAT_1803bbd20[2] != 0)) {
    if ((DAT_1803bbd20 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bbd20 && (*DAT_1803bbd20 = *DAT_1803bbd20 + -1, *plVar9 == 0)))) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    goto LAB_1801c87b9;
  }
  plVar9 = DAT_1803bbd20;
  FUN_180001d00(param_1,DAT_1803bbd20);
/* novels.prompts.inkos.reviser_prompts#46="en" */
  iVar11 = FUN_1803836d0(plVar3,DAT_1803bbf80);
  plVar18 = plVar19;
  if (iVar11 == -1) {
    lVar16 = *(longlong *)(param_1 + 0x60);
    uVar21 = 0xa3;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
/* novels.prompts.inkos.reviser_prompts#3="" */
    plVar25 = DAT_1803bbe28;
    if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#56="【LANGUAGE OVERRIDE】ALL output (FIXED_ISSUES, PATCHES, REVISED_CONTENT, UPDATED_STATE, UPDATED_HOOKS) MUST be in English.\n\n" */
      plVar25 = DAT_1803bbfd0;
    }
    iVar11 = (int)*plVar25 + 1;
    if (iVar11 != 0) {
      *(int *)plVar25 = iVar11;
    }
/* novels.prompts.inkos.reviser_prompts#57="numericalSystem" */
    plVar12 = (longlong *)FUN_180001b80(param_1,plVar2,DAT_1803bbfd8);
    if (plVar12 == (longlong *)0x0) {
      lVar16 = *(longlong *)(param_1 + 0x60);
      uVar21 = 0xa8;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      iVar11 = FUN_180001510(plVar12);
      if (iVar11 == -1) {
        lVar16 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        uVar21 = 0xa8;
      }
      else {
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
/* novels.prompts.inkos.reviser_prompts#3="" */
        plVar20 = DAT_1803bbe28;
        if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#58="\n=== UPDATED_LEDGER ===\n(更新后的完整资源账本)" */
          plVar20 = DAT_1803bbfe0;
        }
        iVar11 = (int)*plVar20 + 1;
        if (iVar11 != 0) {
          *(int *)plVar20 = iVar11;
        }
/* novels.prompts.inkos.reviser_prompts#3="" */
        plVar12 = DAT_1803bbe28;
/* novels.prompts.inkos.reviser_prompts#3="" */
        iVar11 = (int)*DAT_1803bbe28 + 1;
        if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
          *(int *)DAT_1803bbe28 = iVar11;
        }
        local_78 = plVar12;
        iVar11 = FUN_180001510(plVar4);
        if (iVar11 == -1) {
          lVar16 = *(longlong *)(param_1 + 0x60);
          uVar21 = 0xac;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          plVar18 = plVar24;
          if (iVar11 == 0) {
LAB_1801c8e13:
/* novels.prompts.inkos.reviser_prompts#46="en" */
            iVar11 = FUN_1803836d0(plVar3,DAT_1803bbf80);
/* novels.prompts.inkos.reviser_prompts#3="" */
            plVar18 = DAT_1803bbe28;
            if (iVar11 == -1) {
              lVar16 = *(longlong *)(param_1 + 0x60);
              uVar21 = 0xb3;
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar23 = plVar22;
              plVar18 = plVar24;
            }
            else if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#3="" */
              iVar11 = (int)*DAT_1803bbe28 + 1;
              if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                *(int *)DAT_1803bbe28 = iVar11;
              }
/* novels.prompts.inkos.reviser_prompts#24="rewrite-only" */
              iVar11 = FUN_1803836d0(plVar6,DAT_1803bbed0);
/* novels.prompts.inkos.reviser_prompts#70="\n\nROUTING: The reviewer's blocking issues are structural / semantic (character collapse, mainline drift, missing payoff, timeline break, unpaid hook, memo drift, etc.). You MUST output REVISED_CONTENT — do not emit PATCHES, they cannot fix " */
              plVar22 = DAT_1803bc040;
              plVar23 = plVar18;
              if (iVar11 == -1) {
                lVar16 = *(longlong *)(param_1 + 0x60);
                uVar21 = 0xb5;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar18 = plVar24;
              }
              else {
                if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#70="\n\nROUTING: The reviewer's blocking issues are structural / semantic (character collapse, mainline drift, missing payoff, timeline break, unpaid hook, memo drift, etc.). You MUST output REVISED_CONTENT — do not emit PATCHES, they cannot fix " */
                  iVar11 = (int)*DAT_1803bc040 + 1;
                  if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#70="\n\nROUTING: The reviewer's blocking issues are structural / semantic (character collapse, mainline drift, missing payoff, timeline break, unpaid hook, memo drift, etc.). You MUST output REVISED_CONTENT — do not emit PATCHES, they cannot fix " */
                    *(int *)DAT_1803bc040 = iVar11;
                  }
                  plVar23 = plVar22;
                  if (-1 < (int)*plVar18) {
                    *plVar18 = *plVar18 + -1;
                    lVar16 = *plVar18;
joined_r0x0001801c8efc:
                    plVar23 = plVar22;
                    if (lVar16 == 0) {
                      (**(code **)(plVar18[1] + 0x30))(plVar18);
                    }
                  }
                }
                else {
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
                  iVar11 = FUN_1803836d0(plVar6,DAT_1803bbed8);
/* novels.prompts.inkos.reviser_prompts#71="\n\nROUTING: The reviewer's blocking issues are local (wording, paragraph shape, fatigue word, information boundary, knowledge pollution). You MUST output PATCHES only — do not rewrite the whole chapter. If patches are not possible, leave PAT" */
                  plVar22 = DAT_1803bc048;
                  if (iVar11 == -1) {
                    lVar16 = *(longlong *)(param_1 + 0x60);
                    uVar21 = 0xb7;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar18 = plVar24;
                    goto LAB_1801c890a;
                  }
                  if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#71="\n\nROUTING: The reviewer's blocking issues are local (wording, paragraph shape, fatigue word, information boundary, knowledge pollution). You MUST output PATCHES only — do not rewrite the whole chapter. If patches are not possible, leave PAT" */
                    iVar11 = (int)*DAT_1803bc048 + 1;
                    if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#71="\n\nROUTING: The reviewer's blocking issues are local (wording, paragraph shape, fatigue word, information boundary, knowledge pollution). You MUST output PATCHES only — do not rewrite the whole chapter. If patches are not possible, leave PAT" */
                      *(int *)DAT_1803bc048 = iVar11;
                    }
                    plVar23 = plVar22;
                    if (-1 < (int)*plVar18) {
                      *plVar18 = *plVar18 + -1;
                      lVar16 = *plVar18;
                      goto joined_r0x0001801c8efc;
                    }
                  }
                }
                lVar16 = *(longlong *)(param_1 + 0x10);
                plVar22 = *(longlong **)(lVar16 + 0x41df8);
                if (plVar22 == (longlong *)0x0) {
                  plVar22 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,0xd);
                }
                else {
                  *(longlong *)(lVar16 + 0x41df8) = plVar22[3];
                  *(int *)(lVar16 + 0x41e68) = *(int *)(lVar16 + 0x41e68) + -1;
                  *plVar22 = 1;
                }
                plVar18 = plVar22 + -2;
                plVar22[3] = 0;
                plVar22[4] = 0;
                plVar22[5] = 0;
                plVar22[6] = 0;
                plVar22[7] = 0;
                plVar22[8] = 0;
                plVar22[9] = 0;
                plVar22[10] = 0;
                plVar22[0xb] = 0;
                plVar22[0xc] = 0;
                plVar22[0xd] = 0;
                plVar22[0xe] = 0;
                plVar22[0xf] = 0;
                lVar16 = _PyThreadState_GetCurrent();
                lVar16 = *(longlong *)(*(longlong *)(lVar16 + 0x10) + 0xd0);
                puVar8 = *(undefined8 **)(lVar16 + 8);
                *puVar8 = plVar18;
                plVar22[-1] = plVar22[-1] & 3;
                plVar22[-1] = plVar22[-1] | (ulonglong)puVar8;
                *plVar18 = lVar16;
                *(longlong **)(lVar16 + 8) = plVar18;
                iVar11 = (int)*plVar25 + 1;
                if (iVar11 != 0) {
                  *(int *)plVar25 = iVar11;
                }
                plVar22[3] = (longlong)plVar25;
/* novels.prompts.inkos.reviser_prompts#72="You are a professional " */
                piVar7 = DAT_1803bc050;
/* novels.prompts.inkos.reviser_prompts#72="You are a professional " */
                if (*DAT_1803bc050 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#72="You are a professional " */
                  *DAT_1803bc050 = *DAT_1803bc050 + 1;
                }
                plVar22[4] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#73="name" */
                plVar18 = (longlong *)FUN_180001b80(param_1,plVar2,DAT_1803bc058);
                if (plVar18 == (longlong *)0x0) {
                  lVar16 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar21 = 0xb9;
                }
                else {
/* novels.prompts.inkos.reviser_prompts#3="" */
                  lVar16 = FUN_180375520(param_1,plVar18,DAT_1803bbe28);
                  if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                    (**(code **)(plVar18[1] + 0x30))(plVar18);
                  }
                  if (lVar16 == 0) {
                    uVar21 = 0xb9;
                  }
                  else {
                    plVar22[5] = lVar16;
/* novels.prompts.inkos.reviser_prompts#74=" web-fiction revision editor. Fix the chapter according to the review notes." */
                    piVar7 = DAT_1803bc060;
/* novels.prompts.inkos.reviser_prompts#74=" web-fiction revision editor. Fix the chapter according to the review notes." */
                    if (*DAT_1803bc060 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#74=" web-fiction revision editor. Fix the chapter according to the review notes." */
                      *DAT_1803bc060 = *DAT_1803bc060 + 1;
                    }
                    plVar22[6] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                    lVar16 = FUN_180375520(param_1,plVar1,DAT_1803bbe28);
                    if (lVar16 == 0) {
                      uVar21 = 0xb9;
                    }
                    else {
                      plVar22[7] = lVar16;
                      iVar11 = (int)*plVar23 + 1;
                      if (iVar11 != 0) {
                        *(int *)plVar23 = iVar11;
                      }
                      plVar22[8] = (longlong)plVar23;
/* novels.prompts.inkos.reviser_prompts#75="\n\nPATCHES and REVISED_CONTENT serve different problems — choose by problem type, not preference:\n\nPATCHES — for local text issues (wording, dialogue, AI-tell phrases, small continuity errors).\n  Each PATCH quotes the passage to change (a se" */
                      piVar7 = DAT_1803bc068;
/* novels.prompts.inkos.reviser_prompts#75="\n\nPATCHES and REVISED_CONTENT serve different problems — choose by problem type, not preference:\n\nPATCHES — for local text issues (wording, dialogue, AI-tell phrases, small continuity errors).\n  Each PATCH quotes the passage to change (a se" */
                      if (*DAT_1803bc068 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#75="\n\nPATCHES and REVISED_CONTENT serve different problems — choose by problem type, not preference:\n\nPATCHES — for local text issues (wording, dialogue, AI-tell phrases, small continuity errors).\n  Each PATCH quotes the passage to change (a se" */
                        *DAT_1803bc068 = *DAT_1803bc068 + 1;
                      }
                      plVar22[9] = (longlong)piVar7;
                      iVar11 = (int)*local_78 + 1;
                      if (iVar11 != 0) {
                        *(int *)local_78 = iVar11;
                      }
                      plVar22[10] = (longlong)local_78;
/* novels.prompts.inkos.reviser_prompts#76="\n\nIf Critical issues include both local and whole-chapter problems, use REVISED_CONTENT (it addresses everything in one pass).\n\nRevision principles:\n1. Fix root causes — do not apply superficial polish" */
                      piVar7 = DAT_1803bc070;
/* novels.prompts.inkos.reviser_prompts#76="\n\nIf Critical issues include both local and whole-chapter problems, use REVISED_CONTENT (it addresses everything in one pass).\n\nRevision principles:\n1. Fix root causes — do not apply superficial polish" */
                      if (*DAT_1803bc070 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#76="\n\nIf Critical issues include both local and whole-chapter problems, use REVISED_CONTENT (it addresses everything in one pass).\n\nRevision principles:\n1. Fix root causes — do not apply superficial polish" */
                        *DAT_1803bc070 = *DAT_1803bc070 + 1;
                      }
                      plVar22[0xb] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                      lVar16 = FUN_180375520(param_1,plVar5,DAT_1803bbe28);
                      if (lVar16 != 0) {
                        plVar22[0xc] = lVar16;
/* novels.prompts.inkos.reviser_prompts#77="\n2. Hook status must stay in sync with the hooks board. If hook debt briefs are provided, preserve hook payoff scenes\n3. Do not alter the plot direction or core conflicts\n4. Preserve the original language style, rhythm, and pacing — do not " */
                        piVar7 = DAT_1803bc078;
/* novels.prompts.inkos.reviser_prompts#77="\n2. Hook status must stay in sync with the hooks board. If hook debt briefs are provided, preserve hook payoff scenes\n3. Do not alter the plot direction or core conflicts\n4. Preserve the original language style, rhythm, and pacing — do not " */
                        if (*DAT_1803bc078 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#77="\n2. Hook status must stay in sync with the hooks board. If hook debt briefs are provided, preserve hook payoff scenes\n3. Do not alter the plot direction or core conflicts\n4. Preserve the original language style, rhythm, and pacing — do not " */
                          *DAT_1803bc078 = *DAT_1803bc078 + 1;
                        }
                        plVar22[0xd] = (longlong)piVar7;
                        iVar11 = (int)*plVar20 + 1;
                        if (iVar11 != 0) {
                          *(int *)plVar20 = iVar11;
                        }
                        plVar22[0xe] = (longlong)plVar20;
/* novels.prompts.inkos.reviser_prompts#78="\n=== UPDATED_HOOKS ===\n(Full updated hooks board)" */
                        piVar7 = DAT_1803bc080;
/* novels.prompts.inkos.reviser_prompts#78="\n=== UPDATED_HOOKS ===\n(Full updated hooks board)" */
                        if (*DAT_1803bc080 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#78="\n=== UPDATED_HOOKS ===\n(Full updated hooks board)" */
                          *DAT_1803bc080 = *DAT_1803bc080 + 1;
                        }
                        plVar22[0xf] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                        lVar16 = PyUnicode_Join(DAT_1803bbe28,plVar22);
                        if ((-1 < (int)*plVar22) && (*plVar22 = *plVar22 + -1, *plVar22 == 0)) {
                          (**(code **)(plVar22[1] + 0x30))(plVar22);
                        }
                        if (lVar16 != 0) {
LAB_1801c977a:
                          FUN_180001d80(param_1);
                          if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
                            (**(code **)(plVar25[1] + 0x30))(plVar25);
                          }
                          if ((-1 < (int)*plVar20) && (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                            (**(code **)(plVar20[1] + 0x30))(plVar20);
                          }
                          if ((-1 < (int)*local_78) && (*local_78 = *local_78 + -1, *local_78 == 0))
                          {
                            (**(code **)(local_78[1] + 0x30))(local_78);
                          }
                          if (((plVar23 != (longlong *)0x0) && (-1 < (int)*plVar23)) &&
                             (*plVar23 = *plVar23 + -1, *plVar23 == 0)) {
                            (**(code **)(plVar23[1] + 0x30))(plVar23);
                          }
                          if (((plVar19 != (longlong *)0x0) && (-1 < (int)*plVar19)) &&
                             (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
                            (**(code **)(plVar19[1] + 0x30))(plVar19);
                          }
                          if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
                            (**(code **)(plVar2[1] + 0x30))();
                          }
                          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                            (**(code **)(plVar1[1] + 0x30))();
                          }
                          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                            (**(code **)(plVar5[1] + 0x30))();
                          }
                          if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
                            (**(code **)(plVar3[1] + 0x30))();
                          }
                          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
                            (**(code **)(plVar4[1] + 0x30))();
                          }
                          if ((int)*plVar6 < 0) {
                            return lVar16;
                          }
                          *plVar6 = *plVar6 + -1;
                          if (*plVar6 != 0) {
                            return lVar16;
                          }
                          (**(code **)(plVar6[1] + 0x30))();
                          return lVar16;
                        }
                        lVar16 = *(longlong *)(param_1 + 0x60);
                        uVar21 = 0xe6;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar18 = plVar24;
                        goto LAB_1801c890a;
                      }
                      uVar21 = 0xc6;
                    }
                  }
                  lVar16 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                plVar18 = plVar24;
                if ((-1 < (int)*plVar22) && (*plVar22 = *plVar22 + -1, *plVar22 == 0)) {
                  (**(code **)(plVar22[1] + 0x30))(plVar22);
                }
              }
            }
            else {
/* novels.prompts.inkos.reviser_prompts#3="" */
              iVar11 = (int)*DAT_1803bbe28 + 1;
              if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                *(int *)DAT_1803bbe28 = iVar11;
              }
/* novels.prompts.inkos.reviser_prompts#24="rewrite-only" */
              iVar11 = FUN_1803836d0(plVar6,DAT_1803bbed0);
/* novels.prompts.inkos.reviser_prompts#79="\n\n分流指令：reviewer 报告的阻塞问题属于结构/语义错（人设崩、主线偏、爽点缺、时间线错、伏笔未收、memo 偏离等）。你必须输出 REVISED_CONTENT——禁止输出 PATCHES，这类问题不能靠补丁修复。如果无法安全重写，在 FIXED_ISSUES 里说明并留空 REVISED_CONTENT。" */
              plVar24 = DAT_1803bc088;
              if (iVar11 == -1) {
                lVar16 = *(longlong *)(param_1 + 0x60);
                uVar21 = 0xeb;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar23 = plVar22;
              }
              else {
                if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#79="\n\n分流指令：reviewer 报告的阻塞问题属于结构/语义错（人设崩、主线偏、爽点缺、时间线错、伏笔未收、memo 偏离等）。你必须输出 REVISED_CONTENT——禁止输出 PATCHES，这类问题不能靠补丁修复。如果无法安全重写，在 FIXED_ISSUES 里说明并留空 REVISED_CONTENT。" */
                  iVar11 = (int)*DAT_1803bc088 + 1;
                  if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#79="\n\n分流指令：reviewer 报告的阻塞问题属于结构/语义错（人设崩、主线偏、爽点缺、时间线错、伏笔未收、memo 偏离等）。你必须输出 REVISED_CONTENT——禁止输出 PATCHES，这类问题不能靠补丁修复。如果无法安全重写，在 FIXED_ISSUES 里说明并留空 REVISED_CONTENT。" */
                    *(int *)DAT_1803bc088 = iVar11;
                  }
                  plVar19 = plVar24;
                  if (-1 < (int)*plVar18) {
                    *plVar18 = *plVar18 + -1;
                    lVar16 = *plVar18;
joined_r0x0001801c9245:
                    plVar19 = plVar24;
                    if (lVar16 == 0) {
                      (**(code **)(plVar18[1] + 0x30))(plVar18);
                    }
                  }
                }
                else {
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
                  iVar11 = FUN_1803836d0(plVar6,DAT_1803bbed8);
/* novels.prompts.inkos.reviser_prompts#80="\n\n分流指令：reviewer 报告的阻塞问题属于局部错（措辞、段落形状、疲劳词、信息越界、知识污染）。你必须只输出 PATCHES——不要整章改写。如果做不出补丁，留空 PATCHES。" */
                  plVar24 = DAT_1803bc090;
                  if (iVar11 == -1) {
                    lVar16 = *(longlong *)(param_1 + 0x60);
                    uVar21 = 0xed;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar23 = plVar22;
                    goto LAB_1801c890a;
                  }
                  plVar19 = plVar18;
                  if (iVar11 == 1) {
/* novels.prompts.inkos.reviser_prompts#80="\n\n分流指令：reviewer 报告的阻塞问题属于局部错（措辞、段落形状、疲劳词、信息越界、知识污染）。你必须只输出 PATCHES——不要整章改写。如果做不出补丁，留空 PATCHES。" */
                    iVar11 = (int)*DAT_1803bc090 + 1;
                    if (iVar11 != 0) {
/* novels.prompts.inkos.reviser_prompts#80="\n\n分流指令：reviewer 报告的阻塞问题属于局部错（措辞、段落形状、疲劳词、信息越界、知识污染）。你必须只输出 PATCHES——不要整章改写。如果做不出补丁，留空 PATCHES。" */
                      *(int *)DAT_1803bc090 = iVar11;
                    }
                    plVar19 = plVar24;
                    if (-1 < (int)*plVar18) {
                      *plVar18 = *plVar18 + -1;
                      lVar16 = *plVar18;
                      goto joined_r0x0001801c9245;
                    }
                  }
                }
                lVar16 = *(longlong *)(param_1 + 0x10);
                plVar24 = *(longlong **)(lVar16 + 0x41df8);
                if (plVar24 == (longlong *)0x0) {
                  plVar24 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,0xd);
                }
                else {
                  *(longlong *)(lVar16 + 0x41df8) = plVar24[3];
                  *(int *)(lVar16 + 0x41e68) = *(int *)(lVar16 + 0x41e68) + -1;
                  *plVar24 = 1;
                }
                plVar18 = plVar24 + -2;
                plVar24[3] = 0;
                plVar24[4] = 0;
                plVar24[5] = 0;
                plVar24[6] = 0;
                plVar24[7] = 0;
                plVar24[8] = 0;
                plVar24[9] = 0;
                plVar24[10] = 0;
                plVar24[0xb] = 0;
                plVar24[0xc] = 0;
                plVar24[0xd] = 0;
                plVar24[0xe] = 0;
                plVar24[0xf] = 0;
                lVar16 = _PyThreadState_GetCurrent();
                lVar16 = *(longlong *)(*(longlong *)(lVar16 + 0x10) + 0xd0);
                puVar8 = *(undefined8 **)(lVar16 + 8);
                *puVar8 = plVar18;
                plVar24[-1] = plVar24[-1] & 3;
                plVar24[-1] = plVar24[-1] | (ulonglong)puVar8;
                *plVar18 = lVar16;
                *(longlong **)(lVar16 + 8) = plVar18;
                iVar11 = (int)*plVar25 + 1;
                if (iVar11 != 0) {
                  *(int *)plVar25 = iVar11;
                }
                plVar24[3] = (longlong)plVar25;
/* novels.prompts.inkos.reviser_prompts#81="你是一位专业的" */
                piVar7 = DAT_1803bc098;
/* novels.prompts.inkos.reviser_prompts#81="你是一位专业的" */
                if (*DAT_1803bc098 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#81="你是一位专业的" */
                  *DAT_1803bc098 = *DAT_1803bc098 + 1;
                }
                plVar24[4] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#73="name" */
                plVar12 = (longlong *)FUN_180001b80(param_1,plVar2,DAT_1803bc058);
                plVar18 = plVar19;
                if (plVar12 == (longlong *)0x0) {
                  lVar16 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar21 = 0xf0;
                }
                else {
/* novels.prompts.inkos.reviser_prompts#3="" */
                  lVar16 = FUN_180375520(param_1,plVar12,DAT_1803bbe28);
                  if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                  }
                  if (lVar16 == 0) {
                    uVar21 = 0xf0;
                  }
                  else {
                    plVar24[5] = lVar16;
/* novels.prompts.inkos.reviser_prompts#82="网络小说修稿编辑。你的任务是根据审稿意见对章节进行修正。" */
                    piVar7 = DAT_1803bc0a0;
/* novels.prompts.inkos.reviser_prompts#82="网络小说修稿编辑。你的任务是根据审稿意见对章节进行修正。" */
                    if (*DAT_1803bc0a0 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#82="网络小说修稿编辑。你的任务是根据审稿意见对章节进行修正。" */
                      *DAT_1803bc0a0 = *DAT_1803bc0a0 + 1;
                    }
                    plVar24[6] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                    lVar16 = FUN_180375520(param_1,plVar1,DAT_1803bbe28);
                    if (lVar16 == 0) {
                      uVar21 = 0xf0;
                    }
                    else {
                      plVar24[7] = lVar16;
                      iVar11 = (int)*plVar19 + 1;
                      if (iVar11 != 0) {
                        *(int *)plVar19 = iVar11;
                      }
                      plVar24[8] = (longlong)plVar19;
/* novels.prompts.inkos.reviser_prompts#83="\n\nPATCHES 和 REVISED_CONTENT 分别处理不同类型的问题——按问题类型选择，不是按偏好：\n\nPATCHES——处理局部文字问题（措辞、对话、AI痕迹、小的连续性错误）。\n  每个 PATCH 引用要修改的原文段落（一句、一段或多段皆可），给出替换文本。未涉及的内容保持原样。\n\nREVISED_CONTENT——处理全章级问题（字数压缩、结构重组、节奏重排、重大剧情偏离）。\n  输出修正后的完整正文。当 Critical 问题包含字数或结构性问题时，必须使" */
                      piVar7 = DAT_1803bc0a8;
/* novels.prompts.inkos.reviser_prompts#83="\n\nPATCHES 和 REVISED_CONTENT 分别处理不同类型的问题——按问题类型选择，不是按偏好：\n\nPATCHES——处理局部文字问题（措辞、对话、AI痕迹、小的连续性错误）。\n  每个 PATCH 引用要修改的原文段落（一句、一段或多段皆可），给出替换文本。未涉及的内容保持原样。\n\nREVISED_CONTENT——处理全章级问题（字数压缩、结构重组、节奏重排、重大剧情偏离）。\n  输出修正后的完整正文。当 Critical 问题包含字数或结构性问题时，必须使" */
                      if (*DAT_1803bc0a8 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#83="\n\nPATCHES 和 REVISED_CONTENT 分别处理不同类型的问题——按问题类型选择，不是按偏好：\n\nPATCHES——处理局部文字问题（措辞、对话、AI痕迹、小的连续性错误）。\n  每个 PATCH 引用要修改的原文段落（一句、一段或多段皆可），给出替换文本。未涉及的内容保持原样。\n\nREVISED_CONTENT——处理全章级问题（字数压缩、结构重组、节奏重排、重大剧情偏离）。\n  输出修正后的完整正文。当 Critical 问题包含字数或结构性问题时，必须使" */
                        *DAT_1803bc0a8 = *DAT_1803bc0a8 + 1;
                      }
                      plVar24[9] = (longlong)piVar7;
                      iVar11 = (int)*local_78 + 1;
                      if (iVar11 != 0) {
                        *(int *)local_78 = iVar11;
                      }
                      plVar24[10] = (longlong)local_78;
/* novels.prompts.inkos.reviser_prompts#84="\n\n如果 Critical 同时包含局部问题和全章问题，使用 REVISED_CONTENT（一次性解决所有问题）。\n\n修稿原则：\n1. 修根因，不做表面润色" */
                      piVar7 = DAT_1803bc0b0;
/* novels.prompts.inkos.reviser_prompts#84="\n\n如果 Critical 同时包含局部问题和全章问题，使用 REVISED_CONTENT（一次性解决所有问题）。\n\n修稿原则：\n1. 修根因，不做表面润色" */
                      if (*DAT_1803bc0b0 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#84="\n\n如果 Critical 同时包含局部问题和全章问题，使用 REVISED_CONTENT（一次性解决所有问题）。\n\n修稿原则：\n1. 修根因，不做表面润色" */
                        *DAT_1803bc0b0 = *DAT_1803bc0b0 + 1;
                      }
                      plVar24[0xb] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                      lVar16 = FUN_180375520(param_1,plVar5,DAT_1803bbe28);
                      if (lVar16 != 0) {
                        plVar24[0xc] = lVar16;
/* novels.prompts.inkos.reviser_prompts#85="\n2. 伏笔状态必须与伏笔池同步。如果提供了 Hook Debt 简报，必须保留伏笔兑现段落\n3. 不改变剧情走向和核心冲突\n4. 保持原文的语言风格、节奏和呼吸——不要压缩过渡段、不要删掉减速段\n5. 情绪用动作外化（不写\"他感到愤怒\"，写动作）。价值观通过行为传达\n6. 不同角色说话方式必须不同。禁止\"众人齐声惊呼\"\n7. 坏事叠坏事，每层比上一层过分\n\n小目标周期修稿指引：\n- 如果本章应该是\"后效\"阶段但仍在加压，把最密集的冲突段落改写为展示改变的段落——谁失去了什么" */
                        piVar7 = DAT_1803bc0b8;
/* novels.prompts.inkos.reviser_prompts#85="\n2. 伏笔状态必须与伏笔池同步。如果提供了 Hook Debt 简报，必须保留伏笔兑现段落\n3. 不改变剧情走向和核心冲突\n4. 保持原文的语言风格、节奏和呼吸——不要压缩过渡段、不要删掉减速段\n5. 情绪用动作外化（不写\"他感到愤怒\"，写动作）。价值观通过行为传达\n6. 不同角色说话方式必须不同。禁止\"众人齐声惊呼\"\n7. 坏事叠坏事，每层比上一层过分\n\n小目标周期修稿指引：\n- 如果本章应该是\"后效\"阶段但仍在加压，把最密集的冲突段落改写为展示改变的段落——谁失去了什么" */
                        if (*DAT_1803bc0b8 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#85="\n2. 伏笔状态必须与伏笔池同步。如果提供了 Hook Debt 简报，必须保留伏笔兑现段落\n3. 不改变剧情走向和核心冲突\n4. 保持原文的语言风格、节奏和呼吸——不要压缩过渡段、不要删掉减速段\n5. 情绪用动作外化（不写\"他感到愤怒\"，写动作）。价值观通过行为传达\n6. 不同角色说话方式必须不同。禁止\"众人齐声惊呼\"\n7. 坏事叠坏事，每层比上一层过分\n\n小目标周期修稿指引：\n- 如果本章应该是\"后效\"阶段但仍在加压，把最密集的冲突段落改写为展示改变的段落——谁失去了什么" */
                          *DAT_1803bc0b8 = *DAT_1803bc0b8 + 1;
                        }
                        plVar24[0xd] = (longlong)piVar7;
                        iVar11 = (int)*plVar20 + 1;
                        if (iVar11 != 0) {
                          *(int *)plVar20 = iVar11;
                        }
                        plVar24[0xe] = (longlong)plVar20;
/* novels.prompts.inkos.reviser_prompts#86="\n=== UPDATED_HOOKS ===\n(更新后的完整伏笔池)" */
                        piVar7 = DAT_1803bc0c0;
/* novels.prompts.inkos.reviser_prompts#86="\n=== UPDATED_HOOKS ===\n(更新后的完整伏笔池)" */
                        if (*DAT_1803bc0c0 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#86="\n=== UPDATED_HOOKS ===\n(更新后的完整伏笔池)" */
                          *DAT_1803bc0c0 = *DAT_1803bc0c0 + 1;
                        }
                        plVar24[0xf] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                        lVar16 = PyUnicode_Join(DAT_1803bbe28,plVar24);
                        if ((-1 < (int)*plVar24) && (*plVar24 = *plVar24 + -1, *plVar24 == 0)) {
                          (**(code **)(plVar24[1] + 0x30))(plVar24);
                        }
                        if (lVar16 != 0) goto LAB_1801c977a;
                        lVar16 = *(longlong *)(param_1 + 0x60);
                        uVar21 = 0x11d;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        plVar23 = plVar22;
                        goto LAB_1801c890a;
                      }
                      uVar21 = 0xfd;
                    }
                  }
                  lVar16 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                plVar23 = plVar22;
                if ((-1 < (int)*plVar24) && (*plVar24 = *plVar24 + -1, *plVar24 == 0)) {
                  (**(code **)(plVar24[1] + 0x30))(plVar24);
                }
              }
            }
          }
          else {
/* novels.prompts.inkos.reviser_prompts#59="zh" */
            iVar11 = FUN_1803836d0(plVar3,DAT_1803bbfe8);
/* novels.prompts.inkos.reviser_prompts#67="\n  HARD CONSTRAINT: The revised chapter must stay within " */
            piVar10 = DAT_1803bc028;
/* novels.prompts.inkos.reviser_prompts#60="\n  硬性约束：重写后的章节必须控制在 " */
            piVar7 = DAT_1803bbff0;
            if (iVar11 == -1) {
              lVar16 = *(longlong *)(param_1 + 0x60);
              uVar21 = 0xaf;
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar18 = plVar19;
            }
            else {
              lVar16 = *(longlong *)(param_1 + 0x10);
              plVar15 = *(longlong **)(lVar16 + 0x41dc8);
              if (iVar11 == 1) {
                if (plVar15 == (longlong *)0x0) {
                  plVar15 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,7);
                }
                else {
                  *(longlong *)(lVar16 + 0x41dc8) = plVar15[3];
                  *(int *)(lVar16 + 0x41e50) = *(int *)(lVar16 + 0x41e50) + -1;
                  *plVar15 = 1;
                }
                plVar17 = plVar15 + -2;
                plVar15[3] = 0;
                plVar15[4] = 0;
                plVar15[5] = 0;
                plVar15[6] = 0;
                plVar15[7] = 0;
                plVar15[8] = 0;
                plVar15[9] = 0;
                lVar16 = _PyThreadState_GetCurrent();
                lVar16 = *(longlong *)(*(longlong *)(lVar16 + 0x10) + 0xd0);
                puVar8 = *(undefined8 **)(lVar16 + 8);
                *puVar8 = plVar17;
                plVar15[-1] = plVar15[-1] & 3;
                plVar15[-1] = plVar15[-1] | (ulonglong)puVar8;
                *plVar17 = lVar16;
                *(longlong **)(lVar16 + 8) = plVar17;
                if (*piVar7 + 1 != 0) {
                  *piVar7 = *piVar7 + 1;
                }
                plVar15[3] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#61="softMin" */
                plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bbff8);
                if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                  lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                  if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                    (**(code **)(plVar17[1] + 0x30))(plVar17);
                  }
                  if (lVar16 != 0) {
                    plVar15[4] = lVar16;
/* novels.prompts.inkos.reviser_prompts#62="-" */
                    piVar7 = DAT_1803bc000;
/* novels.prompts.inkos.reviser_prompts#62="-" */
                    if (*DAT_1803bc000 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#62="-" */
                      *DAT_1803bc000 = *DAT_1803bc000 + 1;
                    }
                    plVar15[5] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#63="softMax" */
                    plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bc008);
                    if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                      lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                      if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                        (**(code **)(plVar17[1] + 0x30))(plVar17);
                      }
                      if (lVar16 != 0) {
                        plVar15[6] = lVar16;
/* novels.prompts.inkos.reviser_prompts#64=" 字以内（目标 " */
                        piVar7 = DAT_1803bc010;
/* novels.prompts.inkos.reviser_prompts#64=" 字以内（目标 " */
                        if (*DAT_1803bc010 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#64=" 字以内（目标 " */
                          *DAT_1803bc010 = *DAT_1803bc010 + 1;
                        }
                        plVar15[7] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#65="target" */
                        plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bc018);
                        if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                          lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          if (lVar16 != 0) {
                            plVar15[8] = lVar16;
/* novels.prompts.inkos.reviser_prompts#66=" 字，±25%）。这是不可突破的底线。" */
                            piVar7 = DAT_1803bc020;
/* novels.prompts.inkos.reviser_prompts#66=" 字，±25%）。这是不可突破的底线。" */
                            if (*DAT_1803bc020 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#66=" 字，±25%）。这是不可突破的底线。" */
                              *DAT_1803bc020 = *DAT_1803bc020 + 1;
                            }
                            plVar15[9] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                            plVar17 = (longlong *)PyUnicode_Join(DAT_1803bbe28,plVar15);
                            if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                              (**(code **)(plVar15[1] + 0x30))(plVar15);
                            }
                            if (plVar17 != (longlong *)0x0) {
LAB_1801c8df6:
                              local_78 = plVar17;
                              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                              {
                                (**(code **)(plVar12[1] + 0x30))(plVar12);
                              }
                              goto LAB_1801c8e13;
                            }
                            lVar16 = *(longlong *)(param_1 + 0x60);
                            uVar21 = 0xae;
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar23 = plVar22;
                            goto LAB_1801c890a;
                          }
                        }
                      }
                    }
                  }
                }
                uVar21 = 0xae;
              }
              else {
                if (plVar15 == (longlong *)0x0) {
                  plVar15 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,7);
                }
                else {
                  *(longlong *)(lVar16 + 0x41dc8) = plVar15[3];
                  *(int *)(lVar16 + 0x41e50) = *(int *)(lVar16 + 0x41e50) + -1;
                  *plVar15 = 1;
                }
                plVar17 = plVar15 + -2;
                plVar15[3] = 0;
                plVar15[4] = 0;
                plVar15[5] = 0;
                plVar15[6] = 0;
                plVar15[7] = 0;
                plVar15[8] = 0;
                plVar15[9] = 0;
                lVar16 = _PyThreadState_GetCurrent();
                lVar16 = *(longlong *)(*(longlong *)(lVar16 + 0x10) + 0xd0);
                puVar8 = *(undefined8 **)(lVar16 + 8);
                *puVar8 = plVar17;
                plVar15[-1] = plVar15[-1] & 3;
                plVar15[-1] = plVar15[-1] | (ulonglong)puVar8;
                *plVar17 = lVar16;
                *(longlong **)(lVar16 + 8) = plVar17;
                if (*piVar10 + 1 != 0) {
                  *piVar10 = *piVar10 + 1;
                }
                plVar15[3] = (longlong)piVar10;
/* novels.prompts.inkos.reviser_prompts#61="softMin" */
                plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bbff8);
                if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                  lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                  if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                    (**(code **)(plVar17[1] + 0x30))(plVar17);
                  }
                  if (lVar16 != 0) {
                    plVar15[4] = lVar16;
/* novels.prompts.inkos.reviser_prompts#62="-" */
                    piVar7 = DAT_1803bc000;
/* novels.prompts.inkos.reviser_prompts#62="-" */
                    if (*DAT_1803bc000 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#62="-" */
                      *DAT_1803bc000 = *DAT_1803bc000 + 1;
                    }
                    plVar15[5] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#63="softMax" */
                    plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bc008);
                    if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                      lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                      if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                        (**(code **)(plVar17[1] + 0x30))(plVar17);
                      }
                      if (lVar16 != 0) {
                        plVar15[6] = lVar16;
/* novels.prompts.inkos.reviser_prompts#68=" characters (target: " */
                        piVar7 = DAT_1803bc030;
/* novels.prompts.inkos.reviser_prompts#68=" characters (target: " */
                        if (*DAT_1803bc030 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#68=" characters (target: " */
                          *DAT_1803bc030 = *DAT_1803bc030 + 1;
                        }
                        plVar15[7] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#65="target" */
                        plVar17 = (longlong *)FUN_180001b80(param_1,plVar4,DAT_1803bc018);
                        if (plVar17 != (longlong *)0x0) {
/* novels.prompts.inkos.reviser_prompts#3="" */
                          lVar16 = FUN_180375520(param_1,plVar17,DAT_1803bbe28);
                          if ((-1 < (int)*plVar17) && (*plVar17 = *plVar17 + -1, *plVar17 == 0)) {
                            (**(code **)(plVar17[1] + 0x30))(plVar17);
                          }
                          if (lVar16 != 0) {
                            plVar15[8] = lVar16;
/* novels.prompts.inkos.reviser_prompts#69=", ±25%). This is non-negotiable — do not exceed this range." */
                            piVar7 = DAT_1803bc038;
/* novels.prompts.inkos.reviser_prompts#69=", ±25%). This is non-negotiable — do not exceed this range." */
                            if (*DAT_1803bc038 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#69=", ±25%). This is non-negotiable — do not exceed this range." */
                              *DAT_1803bc038 = *DAT_1803bc038 + 1;
                            }
                            plVar15[9] = (longlong)piVar7;
/* novels.prompts.inkos.reviser_prompts#3="" */
                            plVar17 = (longlong *)PyUnicode_Join(DAT_1803bbe28,plVar15);
                            if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                              (**(code **)(plVar15[1] + 0x30))(plVar15);
                            }
                            if (plVar17 != (longlong *)0x0) goto LAB_1801c8df6;
                            lVar16 = *(longlong *)(param_1 + 0x60);
                            uVar21 = 0xb0;
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar23 = plVar22;
                            goto LAB_1801c890a;
                          }
                        }
                      }
                    }
                  }
                }
                uVar21 = 0xb0;
              }
              lVar16 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar18 = plVar19;
              if ((-1 < (int)*plVar15) &&
                 (*plVar15 = *plVar15 + -1, plVar23 = plVar22, plVar18 = plVar24, *plVar15 == 0)) {
                (**(code **)(plVar15[1] + 0x30))(plVar15);
              }
            }
          }
        }
      }
    }
  }
LAB_1801c890a:
  piVar7 = *(int **)(lVar16 + 0x28);
  if (piVar7 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar9,uVar21);
    plVar19 = *(longlong **)(lVar16 + 0x28);
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_1801c951a:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar7 + 6) == plVar9) goto LAB_1801c9563;
    pcVar13 = (code *)FUN_18038e2a0(plVar9,uVar21);
    *(int **)(pcVar13 + 0x10) = piVar7;
    if (*piVar7 + 1 != 0) {
      *piVar7 = *piVar7 + 1;
    }
    plVar19 = *(longlong **)(lVar16 + 0x28);
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_1801c951a;
    pcVar14 = (code *)0x0;
  }
  *(code **)(lVar16 + 0x28) = pcVar14;
  if (((plVar19 != (longlong *)0x0) && (-1 < (int)*plVar19)) &&
     (*plVar19 = *plVar19 + -1, *plVar19 == 0)) {
    (**(code **)(plVar19[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_1801c9563:
  FUN_18036d0a0(plVar9,"ooooooooooo",plVar2,plVar1,plVar5,plVar3,plVar4,plVar6,plVar25,plVar20,
                local_78,plVar23,plVar18);
  plVar19 = DAT_1803bbd20;
  if (plVar9 == DAT_1803bbd20) {
    if ((-1 < (int)*DAT_1803bbd20) && (*DAT_1803bbd20 = *DAT_1803bbd20 + -1, *plVar19 == 0)) {
      (**(code **)(plVar19[1] + 0x30))();
    }
    DAT_1803bbd20 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar25 != (longlong *)0x0) && (-1 < (int)*plVar25)) &&
     (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
    (**(code **)(plVar25[1] + 0x30))(plVar25);
  }
  if (((plVar20 != (longlong *)0x0) && (-1 < (int)*plVar20)) &&
     (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
    (**(code **)(plVar20[1] + 0x30))(plVar20);
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))(local_78);
  }
  if (((plVar23 != (longlong *)0x0) && (-1 < (int)*plVar23)) &&
     (*plVar23 = *plVar23 + -1, *plVar23 == 0)) {
    (**(code **)(plVar23[1] + 0x30))(plVar23);
  }
  if (((plVar18 != (longlong *)0x0) && (-1 < (int)*plVar18)) &&
     (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
    (**(code **)(plVar18[1] + 0x30))(plVar18);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar16;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
