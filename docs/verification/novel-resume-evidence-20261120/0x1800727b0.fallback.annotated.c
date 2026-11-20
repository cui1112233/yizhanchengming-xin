/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800727b0 */

code * FUN_1800727b0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  undefined8 uVar3;
  longlong *plVar4;
  undefined8 uVar5;
  undefined8 uVar6;
  undefined8 uVar7;
  char cVar8;
  int iVar9;
  int iVar10;
  longlong *plVar11;
  longlong *plVar12;
  longlong *plVar13;
  code *pcVar14;
  longlong lVar15;
  code *pcVar16;
  longlong lVar17;
  longlong *plVar18;
  code *pcVar19;
  code *pcVar20;
  code *pcVar21;
  code *pcVar22;
  longlong *local_b8;
  longlong *local_b0;
  code *local_a8;
  longlong local_a0;
  int *local_98;
  code *local_90;
  code *local_88;
  code *local_80;
  code *local_78;
  longlong *local_70;
  longlong *local_68;
  code *local_60;
  longlong *local_58;
  code *local_50;
  longlong *local_48;

  plVar4 = DAT_1803afaf8;
  pcVar20 = (code *)0x0;
  local_48 = (longlong *)*param_3;
  plVar1 = (longlong *)param_3[1];
  local_68 = (longlong *)0x0;
  local_60 = (code *)0x0;
  local_90 = (code *)0x0;
  local_80 = (code *)0x0;
  local_88 = (code *)0x0;
  local_58 = (longlong *)0x0;
  local_50 = (code *)0x0;
  local_70 = (longlong *)0x0;
  local_78 = (code *)0x0;
  if (DAT_1803afaf8 == (longlong *)0x0) {
LAB_18007283e:
    DAT_1803afaf8 = (longlong *)FUN_18036cd10(param_1,DAT_1803afde8,DAT_1803ca828,0x68);
  }
  else if ((1 < *DAT_1803afaf8) || (DAT_1803afaf8[2] != 0)) {
    if ((DAT_1803afaf8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803afaf8 && (*DAT_1803afaf8 = *DAT_1803afaf8 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_18007283e;
  }
  plVar4 = DAT_1803afaf8;
  FUN_180001d00(param_1,DAT_1803afaf8);
  plVar11 = (longlong *)FUN_18037f0c0();
  iVar10 = 0xe6;
  pcVar19 = pcVar20;
  if (plVar11 != (longlong *)0x0) {
    *(undefined4 *)(plVar4 + 5) = 0xe6;
/* novels.api.novels_api#88={"type": "tuple", "items": [{"tag": "a", "fileRange": [3925899, 3925910], "value": "sessionId"}, {"tag": "s", "fileRange": [3925910, 3925911], "value": ""}]} */
    plVar12 = (longlong *)FUN_180378210(param_1,plVar11,DAT_1803b0210);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (plVar12 == (longlong *)0x0) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_180073317;
    }
    plVar11 = (longlong *)FUN_1803749a0(plVar12);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if (plVar11 == (longlong *)0x0) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_180073317;
    }
    plVar12 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (plVar12 == (longlong *)0x0) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_180073317;
    }
    *(undefined4 *)(plVar4 + 5) = 0xe6;
    plVar11 = (longlong *)FUN_180376210(param_1,plVar12);
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if (plVar11 == (longlong *)0x0) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_180073317;
    }
    local_68 = plVar11;
    iVar10 = FUN_180001510(plVar11);
    if (iVar10 == -1) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      iVar10 = 0xe7;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_180073317;
    }
    if (iVar10 != 0) {
      iVar10 = 0xe9;
      *(undefined4 *)(plVar4 + 5) = 0xe9;
/* novels.api.novels_api#33="_state" */
      plVar12 = (longlong *)FUN_18037e560(param_1,local_48,DAT_1803b0058);
      if (plVar12 == (longlong *)0x0) {
        lVar17 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        *(undefined4 *)(plVar4 + 5) = 0xe9;
/* novels.api.novels_api#70="load_session" */
        pcVar21 = (code *)FUN_18037e6c0(param_1,plVar12,DAT_1803b0180,plVar11);
        local_a8 = pcVar21;
        if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        if (pcVar21 == (code *)0x0) {
          lVar17 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          local_60 = pcVar21;
          plVar12 = (longlong *)FUN_18037f0c0();
          if (plVar12 == (longlong *)0x0) {
            lVar17 = *(longlong *)(param_1 + 0x60);
            iVar10 = 0xea;
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            plVar18 = (longlong *)FUN_18037f0c0();
            if (plVar18 == (longlong *)0x0) {
              lVar17 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                (**(code **)(plVar12[1] + 0x30))(plVar12);
              }
LAB_180072ac7:
              iVar10 = 0xea;
            }
            else {
              iVar10 = FUN_180001510(plVar18);
              if (iVar10 == -1) {
                lVar17 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
                if (((int)*plVar18 < 0) || (*plVar18 = *plVar18 + -1, *plVar18 != 0))
                goto LAB_180072ac7;
                (**(code **)(plVar18[1] + 0x30))(plVar18);
                iVar10 = 0xea;
              }
              else {
                if (iVar10 != 1) {
                  if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                    (**(code **)(plVar18[1] + 0x30))(plVar18);
                  }
                  iVar10 = (int)*plVar11 + 1;
                  plVar18 = plVar11;
                  if (iVar10 != 0) {
                    *(int *)plVar11 = iVar10;
                  }
                }
                iVar10 = 0xea;
                *(undefined4 *)(plVar4 + 5) = 0xea;
                plVar13 = (longlong *)FUN_180376770(param_1,plVar12,plVar18);
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
                if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                  (**(code **)(plVar18[1] + 0x30))(plVar18);
                }
                if (plVar13 == (longlong *)0x0) {
                  lVar17 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
                  plVar12 = (longlong *)FUN_18037f0c0();
                  iVar10 = 0xeb;
                  if (plVar12 == (longlong *)0x0) {
                    lVar17 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
                    *(undefined4 *)(plVar4 + 5) = 0xeb;
/* novels.api.novels_api#91={"type": "tuple", "items": [{"tag": "a", "fileRange": [3925973, 3925995], "value": "codexReasoningEffort"}]} */
                    pcVar21 = (code *)FUN_1803770d0(param_1,plVar12,DAT_1803b0228);
                    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                      (**(code **)(plVar12[1] + 0x30))(plVar12);
                    }
                    if (pcVar21 == (code *)0x0) {
                      lVar17 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                    }
                    else {
                      iVar9 = FUN_180001510(pcVar21);
                      if (iVar9 == -1) {
LAB_180072c49:
                        lVar17 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < *(int *)pcVar21) &&
                           (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1,
                           *(longlong *)pcVar21 == 0)) {
                          (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
                        }
                      }
                      else if (iVar9 == 1) {
LAB_180072d00:
                        local_90 = pcVar21;
                        plVar12 = (longlong *)FUN_18037f0c0();
                        iVar10 = 0xec;
                        if (plVar12 == (longlong *)0x0) {
                          lVar17 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          *(undefined4 *)(plVar4 + 5) = 0xec;
/* novels.api.novels_api#93={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926023, 3926046], "value": "claudeReasoningEffort"}]} */
                          pcVar21 = (code *)FUN_1803770d0(param_1,plVar12,DAT_1803b0238);
                          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                          }
                          if (pcVar21 == (code *)0x0) {
                            lVar17 = *(longlong *)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
                            iVar9 = FUN_180001510(pcVar21);
                            if (iVar9 == -1) goto LAB_180072c49;
                            if (iVar9 == 1) {
LAB_180072e0d:
                              pcVar16 = local_90;
                              local_80 = pcVar21;
                              iVar10 = FUN_180001510(local_90);
                              if (iVar10 == -1) {
                                lVar17 = *(longlong *)(param_1 + 0x60);
                                iVar10 = 0xed;
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                              else {
                                if (iVar10 != 1) {
                                  iVar10 = FUN_180001510(pcVar21);
                                  if (iVar10 == -1) {
                                    lVar17 = *(longlong *)(param_1 + 0x60);
                                    iVar10 = 0xed;
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    goto LAB_180073317;
                                  }
/* novels.api.novels_api#17="" */
                                  pcVar16 = DAT_1803affd8;
                                  if (iVar10 == 1) {
                                    pcVar16 = pcVar21;
                                  }
                                }
                                plVar12 = (longlong *)FUN_1803749a0(pcVar16);
                                if (plVar12 == (longlong *)0x0) {
                                  lVar17 = *(longlong *)(param_1 + 0x60);
                                  iVar10 = 0xed;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                else {
                                  plVar18 = (longlong *)FUN_18037f0c0();
                                  if ((-1 < (int)*plVar12) &&
                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                                  }
                                  iVar10 = 0xed;
                                  if (plVar18 == (longlong *)0x0) {
                                    lVar17 = *(longlong *)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                  }
                                  else {
                                    *(undefined4 *)(plVar4 + 5) = 0xed;
                                    pcVar21 = (code *)FUN_180376210(param_1,plVar18);
                                    if ((-1 < (int)*plVar18) &&
                                       (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                      (**(code **)(plVar18[1] + 0x30))(plVar18);
                                    }
                                    if (pcVar21 == (code *)0x0) {
                                      lVar17 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                    }
                                    else {
                                      iVar9 = FUN_180001510(pcVar21);
                                      if (iVar9 == -1) {
LAB_180072f1a:
                                        lVar17 = *(longlong *)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        if ((-1 < *(int *)pcVar21) &&
                                           (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1,
                                           *(longlong *)pcVar21 == 0)) {
                                          lVar15 = *(longlong *)(pcVar21 + 8);
LAB_18007330d:
                                          (**(code **)(lVar15 + 0x30))(pcVar21);
                                        }
                                      }
                                      else {
                                        if (iVar9 != 1) {
                                          if ((-1 < *(int *)pcVar21) &&
                                             (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1,
                                             *(longlong *)pcVar21 == 0)) {
                                            (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))
                                                      (pcVar21);
                                          }
                                          local_88 = _Py_NoneStruct_exref;
                                          pcVar21 = local_88;
                                          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                                            *(int *)_Py_NoneStruct_exref =
                                                 *(int *)_Py_NoneStruct_exref + 1;
                                          }
                                        }
                                        local_88 = pcVar21;
                                        pcVar21 = local_88;
                                        plVar12 = (longlong *)FUN_18037f0c0();
                                        if (plVar12 == (longlong *)0x0) {
                                          lVar17 = *(longlong *)(param_1 + 0x60);
                                          iVar10 = 0xee;
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                        }
                                        else {
                                          plVar18 = (longlong *)FUN_18037f0c0();
                                          if (plVar18 == (longlong *)0x0) {
                                            lVar17 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            if ((-1 < (int)*plVar12) &&
                                               (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                                            }
                                            iVar10 = 0xee;
                                          }
                                          else {
                                            iVar10 = 0xee;
                                            *(undefined4 *)(plVar4 + 5) = 0xee;
                                            local_b8 = plVar11;
                                            local_b0 = plVar18;
                                            local_a8 = pcVar21;
                                            plVar13 = (longlong *)
                                                      FUN_18037d650(param_1,plVar12,&local_b8,
/* novels.api.novels_api#49={"type": "tuple", "items": [{"tag": "a", "fileRange": [3925265, 3925281], "value": "channel_effort"}]} */
                                                                    &local_a8,DAT_1803b00d8);
                                            if ((-1 < (int)*plVar12) &&
                                               (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                              (**(code **)(plVar12[1] + 0x30))(plVar12);
                                            }
                                            if ((-1 < (int)*plVar18) &&
                                               (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
                                              (**(code **)(plVar18[1] + 0x30))(plVar18);
                                            }
                                            if (plVar13 == (longlong *)0x0) {
                                              lVar17 = *(longlong *)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                            }
                                            else {
                                              local_58 = plVar13;
                                              pcVar16 = (code *)FUN_18037f0c0();
                                              if (pcVar16 == (code *)0x0) {
                                                lVar17 = *(longlong *)(param_1 + 0x60);
                                                iVar10 = 0xef;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                              }
                                              else {
                                                plVar12 = (longlong *)FUN_18037f0c0();
                                                if (plVar12 == (longlong *)0x0) {
                                                  lVar17 = *(longlong *)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  if ((-1 < *(int *)pcVar16) &&
                                                     (*(longlong *)pcVar16 =
                                                           *(longlong *)pcVar16 + -1,
                                                     *(longlong *)pcVar16 == 0)) {
                                                    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))
                                                              (pcVar16);
                                                  }
                                                  iVar10 = 0xf2;
                                                }
                                                else {
                                                  iVar10 = 0xf2;
                                                  *(undefined4 *)(plVar4 + 5) = 0xf2;
                                                  pcVar14 = (code *)FUN_1803770d0(param_1,plVar12,
/* novels.api.novels_api#97={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926120, 3926127], "value": "model"}]} */
                                                                                  DAT_1803b0258);
                                                  if ((-1 < (int)*plVar12) &&
                                                     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                    (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                  }
                                                  pcVar21 = pcVar16;
                                                  if (pcVar14 == (code *)0x0) goto LAB_180072f1a;
                                                  iVar10 = 0xef;
                                                  *(undefined4 *)(plVar4 + 5) = 0xef;
                                                  local_b8 = plVar11;
                                                  local_b0 = plVar1;
                                                  local_a8 = pcVar14;
                                                  pcVar21 = (code *)FUN_18037d650(param_1,pcVar16,
                                                                                  &local_b8,
                                                                                  &local_a8,
/* novels.api.novels_api#97={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926120, 3926127], "value": "model"}]} */
                                                                                  DAT_1803b0258);
                                                  if ((-1 < *(int *)pcVar16) &&
                                                     (*(longlong *)pcVar16 =
                                                           *(longlong *)pcVar16 + -1,
                                                     *(longlong *)pcVar16 == 0)) {
                                                    (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))
                                                              (pcVar16);
                                                  }
                                                  if ((-1 < *(int *)pcVar14) &&
                                                     (*(longlong *)pcVar14 =
                                                           *(longlong *)pcVar14 + -1,
                                                     *(longlong *)pcVar14 == 0)) {
                                                    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))
                                                              (pcVar14);
                                                  }
                                                  if (pcVar21 == (code *)0x0) {
                                                    lVar17 = *(longlong *)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                  }
                                                  else {
                                                    local_50 = pcVar21;
                                                    plVar11 = (longlong *)FUN_18037f0c0();
                                                    iVar10 = 0xf4;
                                                    if (plVar11 == (longlong *)0x0) {
                                                      lVar17 = *(longlong *)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                    }
                                                    else {
                                                      *(undefined4 *)(plVar4 + 5) = 0xf4;
                                                      plVar12 = (longlong *)
                                                                FUN_1803770d0(param_1,plVar11,
/* novels.api.novels_api#39={"type": "tuple", "items": [{"tag": "a", "fileRange": [3925049, 3925058], "value": "success"}]} */
                                                                              DAT_1803b0088);
                                                      if ((-1 < (int)*plVar11) &&
                                                         (*plVar11 = *plVar11 + -1, *plVar11 == 0))
                                                      {
                                                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                      }
                                                      if (plVar12 == (longlong *)0x0) {
                                                        lVar17 = *(longlong *)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                      }
                                                      else {
                                                        iVar9 = FUN_180001510(plVar12);
                                                        if ((-1 < (int)*plVar12) &&
                                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)
                                                           ) {
                                                          (**(code **)(plVar12[1] + 0x30))(plVar12);
                                                        }
/* novels.api.novels_api#58="success" */
                                                        uVar3 = DAT_1803b0120;
                                                        pcVar16 = _Py_TrueStruct_exref;
                                                        if (iVar9 == -1) {
                                                          lVar17 = *(longlong *)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                        }
                                                        else {
                                                          if (iVar9 == 0) {
                                                            if (*(int *)pcVar21 + 1 != 0) {
                                                              *(int *)pcVar21 = *(int *)pcVar21 + 1;
                                                            }
LAB_180073c0e:
                                                            FUN_180001d80(param_1);
                                                            pcVar16 = local_88;
                                                            pcVar14 = local_80;
                                                            pcVar22 = local_90;
                                                            goto LAB_180073c25;
                                                          }
                                                          pcVar21 = (code *)_PyDict_NewPresized(2);
                                                          PyDict_SetItem(pcVar21,uVar3,pcVar16);
/* novels.api.novels_api#8="data" */
                                                          local_b8 = DAT_1803aff90;
                                                          plVar11 = (longlong *)FUN_18037f0c0();
                                                          iVar10 = 0xf6;
                                                          if (plVar11 != (longlong *)0x0) {
                                                            *(undefined4 *)(plVar4 + 5) = 0xf6;
                                                            plVar12 = (longlong *)
                                                                      FUN_1803770d0(param_1,plVar11,
/* novels.api.novels_api#98={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926129, 3926135], "value": "data"}]} */
                                                                                    DAT_1803b0260);
                                                            if ((-1 < (int)*plVar11) &&
                                                               (*plVar11 = *plVar11 + -1,
                                                               *plVar11 == 0)) {
                                                              (**(code **)(plVar11[1] + 0x30))
                                                                        (plVar11);
                                                            }
                                                            if (plVar12 != (longlong *)0x0) {
                                                              PyDict_SetItem(pcVar21,local_b8,
                                                                             plVar12);
                                                              if ((-1 < (int)*plVar12) &&
                                                                 (*plVar12 = *plVar12 + -1,
                                                                 *plVar12 == 0)) {
                                                                (**(code **)(plVar12[1] + 0x30))
                                                                          (plVar12);
                                                              }
                                                              goto LAB_180073c0e;
                                                            }
                                                          }
                                                          lVar17 = *(longlong *)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          if ((-1 < *(int *)pcVar21) &&
                                                             (*(longlong *)pcVar21 =
                                                                   *(longlong *)pcVar21 + -1,
                                                             *(longlong *)pcVar21 == 0)) {
                                                            lVar15 = *(longlong *)(pcVar21 + 8);
                                                            goto LAB_18007330d;
                                                          }
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                            else {
                              if ((-1 < *(int *)pcVar21) &&
                                 (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1,
                                 *(longlong *)pcVar21 == 0)) {
                                (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
                              }
                              plVar12 = (longlong *)FUN_18037f0c0();
                              if (plVar12 == (longlong *)0x0) {
                                lVar17 = *(longlong *)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                              else {
                                *(undefined4 *)(plVar4 + 5) = 0xec;
/* novels.api.novels_api#94={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926048, 3926073], "value": "claude_reasoning_effort"}]} */
                                pcVar21 = (code *)FUN_1803770d0(param_1,plVar12,DAT_1803b0240);
                                if ((-1 < (int)*plVar12) &&
                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                                }
                                if (pcVar21 != (code *)0x0) goto LAB_180072e0d;
                                lVar17 = *(longlong *)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                            }
                          }
                        }
                      }
                      else {
                        if ((-1 < *(int *)pcVar21) &&
                           (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1,
                           *(longlong *)pcVar21 == 0)) {
                          (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
                        }
                        plVar12 = (longlong *)FUN_18037f0c0();
                        if (plVar12 == (longlong *)0x0) {
                          lVar17 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          *(undefined4 *)(plVar4 + 5) = 0xeb;
/* novels.api.novels_api#92={"type": "tuple", "items": [{"tag": "a", "fileRange": [3925997, 3926021], "value": "codex_reasoning_effort"}]} */
                          pcVar21 = (code *)FUN_1803770d0(param_1,plVar12,DAT_1803b0230);
                          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                            (**(code **)(plVar12[1] + 0x30))(plVar12);
                          }
                          if (pcVar21 != (code *)0x0) goto LAB_180072d00;
                          lVar17 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
      goto LAB_180073317;
    }
/* novels.api.novels_api#89={"type": "dict", "keys": [{"tag": "a", "fileRange": [3925913, 3925922], "value": "success"}, {"tag": "a", "fileRange": [3925922, 3925931], "value": "message"}], "values": [{"tag": "F", "fileRange": [3925931, 3925932], "value": false}, {"tag": "u", "fileRange": [3925932, 3925956], "value": "sessionId 不能为空"}]} */
    pcVar21 = (code *)FUN_18036fb80(param_1,DAT_1803b0218);
    FUN_180001d80(param_1);
    pcVar16 = pcVar20;
    pcVar14 = pcVar20;
    pcVar22 = pcVar20;
LAB_180073c25:
    if ((-1 < (int)*local_68) && (*local_68 = *local_68 + -1, *local_68 == 0)) {
      (**(code **)(local_68[1] + 0x30))(local_68);
    }
LAB_180073c44:
    if (((local_60 != (code *)0x0) && (-1 < *(int *)local_60)) &&
       (*(longlong *)local_60 = *(longlong *)local_60 + -1, *(longlong *)local_60 == 0)) {
      (**(code **)(*(longlong *)(local_60 + 8) + 0x30))();
    }
    if (((pcVar22 != (code *)0x0) && (-1 < *(int *)pcVar22)) &&
       (*(longlong *)pcVar22 = *(longlong *)pcVar22 + -1, *(longlong *)pcVar22 == 0)) {
      (**(code **)(*(longlong *)(pcVar22 + 8) + 0x30))(pcVar22);
    }
    if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
       (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
      (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
    }
    if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
       (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
      (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
    }
    if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
       (*local_58 = *local_58 + -1, *local_58 == 0)) {
      (**(code **)(local_58[1] + 0x30))();
    }
    if (((local_50 != (code *)0x0) && (-1 < *(int *)local_50)) &&
       (*(longlong *)local_50 = *(longlong *)local_50 + -1, *(longlong *)local_50 == 0)) {
      (**(code **)(*(longlong *)(local_50 + 8) + 0x30))();
    }
    if (((local_70 != (longlong *)0x0) && (-1 < (int)*local_70)) &&
       (*local_70 = *local_70 + -1, *local_70 == 0)) {
      (**(code **)(local_70[1] + 0x30))();
    }
    if (((pcVar20 != (code *)0x0) && (-1 < *(int *)pcVar20)) &&
       (*(longlong *)pcVar20 = *(longlong *)pcVar20 + -1, *(longlong *)pcVar20 == 0)) {
      (**(code **)(*(longlong *)(pcVar20 + 8) + 0x30))(pcVar20);
    }
    if (((pcVar19 != (code *)0x0) && (-1 < *(int *)pcVar19)) &&
       (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1, *(longlong *)pcVar19 == 0)) {
      (**(code **)(*(longlong *)(pcVar19 + 8) + 0x30))(pcVar19);
    }
    if ((-1 < (int)*local_48) && (*local_48 = *local_48 + -1, *local_48 == 0)) {
      (**(code **)(local_48[1] + 0x30))();
    }
    if ((int)*plVar1 < 0) {
      return pcVar21;
    }
    *plVar1 = *plVar1 + -1;
    if (*plVar1 != 0) {
      return pcVar21;
    }
    (**(code **)(plVar1[1] + 0x30))();
    return pcVar21;
  }
  lVar17 = *(longlong *)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
LAB_180073317:
  plVar11 = local_68;
  local_98 = *(int **)(param_1 + 0xf8);
  local_b8 = local_68;
  local_a0 = 0;
  if ((local_98 != (int *)0x0) && (*local_98 + 1 != 0)) {
    *local_98 = *local_98 + 1;
  }
  piVar2 = *(int **)(lVar17 + 0x28);
  pcVar21 = pcVar20;
  if (piVar2 == (int *)0x0) {
    pcVar16 = (code *)FUN_18038e2a0(plVar4,iVar10);
    plVar12 = *(longlong **)(lVar17 + 0x28);
    if (pcVar16 != _Py_NoneStruct_exref) {
LAB_180073396:
      pcVar21 = pcVar16;
      if ((pcVar16 != (code *)0x0) && (*(int *)pcVar16 + 1 != 0)) {
        *(int *)pcVar16 = *(int *)pcVar16 + 1;
      }
    }
LAB_1800733a4:
    *(code **)(lVar17 + 0x28) = pcVar21;
    if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
       (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))();
    }
    if (((pcVar16 != (code *)0x0) && (-1 < *(int *)pcVar16)) &&
       (*(longlong *)pcVar16 = *(longlong *)pcVar16 + -1, *(longlong *)pcVar16 == 0)) {
      (**(code **)(*(longlong *)(pcVar16 + 8) + 0x30))(pcVar16);
    }
  }
  else if (iVar10 != 0) {
    pcVar16 = (code *)FUN_18038e2a0(plVar4,iVar10);
    *(int **)(pcVar16 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar12 = *(longlong **)(lVar17 + 0x28);
    if (pcVar16 != _Py_NoneStruct_exref) goto LAB_180073396;
    goto LAB_1800733a4;
  }
  plVar12 = *(longlong **)(param_1 + 0xf8);
  *(longlong *)(param_1 + 0xf8) = lVar17;
  if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
     (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
    (**(code **)(plVar12[1] + 0x30))();
  }
  uVar3 = *(undefined8 *)(*(longlong *)(param_1 + 0xf8) + 8);
  lVar17 = FUN_180066230(param_1);
  if (lVar17 == 0) {
/* novels.api.novels_api#99="OperationCancelledError" */
    FUN_180373340(param_1,&local_a0,DAT_1803b0268);
    pcVar21 = (code *)0xf7;
    lVar17 = local_a0;
  }
  else {
    iVar10 = FUN_18000ce90(param_1,uVar3,lVar17);
    if (iVar10 == -1) {
      lVar17 = *(longlong *)(param_1 + 0x60);
      pcVar21 = (code *)0xf7;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar12 = *(longlong **)(param_1 + 0xf8);
      if (iVar10 == 0) {
        iVar10 = FUN_18000ce90(param_1,plVar12[1],*(undefined8 *)PyExc_Exception_exref);
        if (iVar10 == 0) {
          cVar8 = FUN_180373290(param_1,&local_a0);
          pcVar21 = (code *)0xe5;
          if (cVar8 != '\0') {
            pcVar21 = pcVar20;
          }
          lVar15 = *(longlong *)(local_a0 + 0x28);
          lVar17 = local_a0;
          if ((lVar15 != 0) && (*(longlong **)(lVar15 + 0x18) == plVar4)) {
            *(undefined4 *)(plVar4 + 5) = *(undefined4 *)(lVar15 + 0x24);
          }
        }
        else {
          plVar12 = *(longlong **)(param_1 + 0xf8);
          iVar10 = (int)*plVar12 + 1;
          if (iVar10 != 0) {
            *(int *)plVar12 = iVar10;
          }
          uVar7 = DAT_1803ca830;
/* novels.api.novels_api#103=2 */
          uVar6 = DAT_1803b0288;
/* novels.api.novels_api#102={"type": "tuple", "items": [{"tag": "a", "fileRange": [3926190, 3926210], "value": "friendly_llm_error"}, {"tag": "a", "fileRange": [3926210, 3926234], "value": "is_transient_llm_error"}]} */
          uVar5 = DAT_1803b0280;
/* novels.api.novels_api#101="utils.llm_retry" */
          uVar3 = DAT_1803b0278;
          pcVar19 = _Py_NoneStruct_exref;
          pcVar21 = (code *)0xfa;
          *(undefined4 *)(plVar4 + 5) = 0xfa;
          plVar11 = (longlong *)FUN_1803717e0(param_1,uVar3,uVar7,pcVar19,uVar5,uVar6);
          if (plVar11 == (longlong *)0x0) goto LAB_180073770;
          if (((code *)plVar11[1] == PyModule_Type_exref) ||
             (iVar10 = PyType_IsSubtype(), iVar10 != 0)) {
            plVar18 = (longlong *)
/* novels.api.novels_api#104="friendly_llm_error" */
                      FUN_180371d80(param_1,plVar11,DAT_1803ca830,DAT_1803b0290,DAT_1803ab250);
          }
          else {
/* novels.api.novels_api#104="friendly_llm_error" */
            plVar18 = (longlong *)FUN_180371890(param_1,plVar11,DAT_1803b0290);
          }
          if (plVar18 == (longlong *)0x0) {
LAB_1800736aa:
            lVar17 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
          }
          else {
            local_70 = plVar18;
            if (((code *)plVar11[1] == PyModule_Type_exref) ||
               (iVar10 = PyType_IsSubtype(), iVar10 != 0)) {
/* novels.api.novels_api#105="is_transient_llm_error" */
              pcVar19 = (code *)FUN_180371d80(param_1,plVar11,DAT_1803ca830,DAT_1803b0298,
                                              DAT_1803ab250);
            }
            else {
/* novels.api.novels_api#105="is_transient_llm_error" */
              pcVar19 = (code *)FUN_180371890(param_1,plVar11,DAT_1803b0298);
            }
            if (pcVar19 == (code *)0x0) goto LAB_1800736aa;
            local_78 = pcVar19;
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            pcVar21 = (code *)0xfb;
            *(undefined4 *)(plVar4 + 5) = 0xfb;
            plVar11 = (longlong *)FUN_180376770(param_1,pcVar19,plVar12);
            if (plVar11 != (longlong *)0x0) {
              iVar10 = FUN_180001510(plVar11);
              if (iVar10 == -1) goto LAB_1800736aa;
              if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                (**(code **)(plVar11[1] + 0x30))(plVar11);
              }
              if (iVar10 == 0) {
                pcVar19 = (code *)FUN_1803749a0(plVar12);
              }
              else {
                *(undefined4 *)(plVar4 + 5) = 0xfb;
                pcVar19 = (code *)FUN_180376770(param_1,plVar18,plVar12);
              }
/* novels.api.novels_api#58="success" */
              uVar3 = DAT_1803b0120;
              pcVar16 = _Py_FalseStruct_exref;
              if (pcVar19 != (code *)0x0) {
                pcVar21 = (code *)_PyDict_NewPresized(2);
                PyDict_SetItem(pcVar21,uVar3,pcVar16);
/* novels.api.novels_api#59="message" */
                PyDict_SetItem(pcVar21,DAT_1803b0128,pcVar19);
                plVar11 = local_b8;
                if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                  plVar11 = local_b8;
                }
                goto LAB_180073809;
              }
            }
LAB_180073770:
            lVar17 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
          }
        }
      }
      else {
        iVar10 = (int)*plVar12 + 1;
        if (iVar10 != 0) {
          *(int *)plVar12 = iVar10;
        }
/* novels.api.novels_api#58="success" */
        uVar3 = DAT_1803b0120;
        pcVar16 = _Py_FalseStruct_exref;
        pcVar21 = (code *)_PyDict_NewPresized(3);
        PyDict_SetItem(pcVar21,uVar3,pcVar16);
/* novels.api.novels_api#100="cancelled" */
        PyDict_SetItem(pcVar21,DAT_1803b0270,_Py_TrueStruct_exref);
/* novels.api.novels_api#59="message" */
        uVar3 = DAT_1803b0128;
        plVar18 = (longlong *)FUN_1803749a0(plVar12);
        if (plVar18 != (longlong *)0x0) {
          PyDict_SetItem(pcVar21,uVar3,plVar18);
          if ((-1 < (int)*plVar18) && (*plVar18 = *plVar18 + -1, *plVar18 == 0)) {
            (**(code **)(plVar18[1] + 0x30))(plVar18);
          }
          if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
             (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
            pcVar19 = (code *)0x0;
          }
LAB_180073809:
          pcVar20 = local_78;
          plVar4 = *(longlong **)(param_1 + 0xf8);
          *(int **)(param_1 + 0xf8) = local_98;
          if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) &&
             (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
            (**(code **)(plVar4[1] + 0x30))();
          }
          FUN_180001d80(param_1);
          pcVar16 = local_88;
          pcVar14 = local_80;
          pcVar22 = local_90;
          local_68 = plVar11;
          if (plVar11 == (longlong *)0x0) goto LAB_180073c44;
          goto LAB_180073c25;
        }
        lVar17 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < *(int *)pcVar21) &&
           (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
          (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
        }
        if (((plVar12 != (longlong *)0x0) && (-1 < (int)*plVar12)) &&
           (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
          (**(code **)(plVar12[1] + 0x30))(plVar12);
        }
        pcVar21 = (code *)0xf8;
      }
    }
  }
  plVar11 = *(longlong **)(param_1 + 0xf8);
  *(int **)(param_1 + 0xf8) = local_98;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  piVar2 = *(int **)(lVar17 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar19 = (code *)FUN_18038e2a0(plVar4,pcVar21);
    plVar11 = *(longlong **)(lVar17 + 0x28);
    if (pcVar19 != _Py_NoneStruct_exref) {
LAB_18007393f:
      pcVar20 = pcVar19;
      pcVar19 = pcVar20;
      if ((pcVar20 != (code *)0x0) && (*(int *)pcVar20 + 1 != 0)) {
        *(int *)pcVar20 = *(int *)pcVar20 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar4) goto LAB_180073988;
    pcVar19 = (code *)FUN_18038e2a0(plVar4,pcVar21);
    *(int **)(pcVar19 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar11 = *(longlong **)(lVar17 + 0x28);
    if (pcVar19 != _Py_NoneStruct_exref) goto LAB_18007393f;
  }
  *(code **)(lVar17 + 0x28) = pcVar20;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  if (((pcVar19 != (code *)0x0) && (-1 < *(int *)pcVar19)) &&
     (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1, *(longlong *)pcVar19 == 0)) {
    (**(code **)(*(longlong *)(pcVar19 + 8) + 0x30))(pcVar19);
  }
LAB_180073988:
  pcVar21 = local_80;
  pcVar19 = local_88;
  pcVar20 = local_90;
  FUN_18036d0a0(plVar4,"ooooooooooooo",local_48,plVar1,local_68,local_60,local_90,local_80,local_88,
                local_58,local_50,0,local_70,local_78,0);
  plVar11 = DAT_1803afaf8;
  if (plVar4 == DAT_1803afaf8) {
    if ((-1 < (int)*DAT_1803afaf8) && (*DAT_1803afaf8 = *DAT_1803afaf8 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    DAT_1803afaf8 = (longlong *)0x0;
  }
  lVar15 = **(longlong **)(param_1 + 0x38);
  plVar4 = *(longlong **)(lVar15 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar15 + 8);
  *(undefined4 *)(plVar4 + 8) = 0xffffffff;
  plVar11 = (longlong *)plVar4[2];
  if (((plVar11 != (longlong *)0x0) && (plVar4[2] = 0, -1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  plVar4[10] = 0;
  if (((local_68 != (longlong *)0x0) && (-1 < (int)*local_68)) &&
     (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
  }
  if (((local_60 != (code *)0x0) && (-1 < *(int *)local_60)) &&
     (*(longlong *)local_60 = *(longlong *)local_60 + -1, *(longlong *)local_60 == 0)) {
    (**(code **)(*(longlong *)(local_60 + 8) + 0x30))();
  }
  if (((pcVar20 != (code *)0x0) && (-1 < *(int *)pcVar20)) &&
     (*(longlong *)pcVar20 = *(longlong *)pcVar20 + -1, *(longlong *)pcVar20 == 0)) {
    (**(code **)(*(longlong *)(pcVar20 + 8) + 0x30))(pcVar20);
  }
  if (((pcVar21 != (code *)0x0) && (-1 < *(int *)pcVar21)) &&
     (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
    (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
  }
  if (((pcVar19 != (code *)0x0) && (-1 < *(int *)pcVar19)) &&
     (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1, *(longlong *)pcVar19 == 0)) {
    (**(code **)(*(longlong *)(pcVar19 + 8) + 0x30))(pcVar19);
  }
  if (((local_58 != (longlong *)0x0) && (-1 < (int)*local_58)) &&
     (*local_58 = *local_58 + -1, *local_58 == 0)) {
    (**(code **)(local_58[1] + 0x30))();
  }
  if (((local_50 != (code *)0x0) && (-1 < *(int *)local_50)) &&
     (*(longlong *)local_50 = *(longlong *)local_50 + -1, *(longlong *)local_50 == 0)) {
    (**(code **)(*(longlong *)(local_50 + 8) + 0x30))();
  }
  if (((local_70 != (longlong *)0x0) && (-1 < (int)*local_70)) &&
     (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  if (((local_78 != (code *)0x0) && (-1 < *(int *)local_78)) &&
     (*(longlong *)local_78 = *(longlong *)local_78 + -1, *(longlong *)local_78 == 0)) {
    (**(code **)(*(longlong *)(local_78 + 8) + 0x30))();
  }
  if ((-1 < (int)*local_48) && (*local_48 = *local_48 + -1, *local_48 == 0)) {
    (**(code **)(local_48[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar17;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
