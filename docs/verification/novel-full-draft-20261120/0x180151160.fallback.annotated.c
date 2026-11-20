/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180151160 */

longlong FUN_180151160(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  longlong lVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  undefined8 uVar12;
  longlong *plVar13;
  longlong *local_res20;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;

  plVar3 = DAT_1803b6bd0;
  plVar1 = (longlong *)*param_3;
  plVar13 = (longlong *)0x0;
  local_res20 = DAT_180391d20;
  if (DAT_1803b6bd0 == (longlong *)0x0) {
LAB_1801511c0:
    DAT_1803b6bd0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6ab0,DAT_1803ca758,0x10);
  }
  else if ((1 < *DAT_1803b6bd0) || (DAT_1803b6bd0[2] != 0)) {
    if ((DAT_1803b6bd0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6bd0 && (*DAT_1803b6bd0 = *DAT_1803b6bd0 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1801511c0;
  }
  plVar3 = DAT_1803b6bd0;
  FUN_180001d00(param_1,DAT_1803b6bd0);
  lVar4 = FUN_18014c500(param_1);
  if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
    FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
    uVar12 = 0x77;
    plVar8 = local_res20;
    plVar13 = (longlong *)0x0;
  }
  else {
    plVar5 = (longlong *)FUN_18037f0c0();
/* novels.pipeline.short_fiction_parsers#4="" */
    plVar7 = DAT_1803b6c50;
    if (plVar5 == (longlong *)0x0) {
      plVar8 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      uVar12 = 0x77;
    }
    else {
/* novels.pipeline.short_fiction_parsers#30="^```(?:md|markdown)?\\s*" */
      local_70 = DAT_1803b6d20;
      lVar4 = FUN_18014c500(param_1);
      if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
        FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        uVar12 = 0x77;
        plVar8 = local_res20;
      }
      else {
        plVar6 = (longlong *)FUN_18037f0c0();
        if (plVar6 == (longlong *)0x0) {
          plVar8 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
          uVar12 = 0x77;
        }
        else {
          local_58 = local_70;
          uVar12 = 0x77;
          local_50 = plVar7;
          *(undefined4 *)(plVar3 + 5) = 0x77;
          local_70 = plVar6;
          local_48 = plVar1;
/* novels.pipeline.short_fiction_parsers#31={"type": "tuple", "items": [{"tag": "a", "fileRange": [3999233, 3999240], "value": "flags"}]} */
          plVar7 = (longlong *)FUN_18037da70(param_1,plVar5,&local_58,&local_70,DAT_1803b6d28);
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if (plVar7 == (longlong *)0x0) {
            plVar8 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar13 = (longlong *)0x0;
          }
          else {
            lVar4 = FUN_18014c500(param_1);
            plVar13 = plVar7;
            if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
              FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
              uVar12 = 0x78;
              plVar8 = local_res20;
            }
            else {
              plVar5 = (longlong *)FUN_18037f0c0();
              if (plVar5 == (longlong *)0x0) {
                uVar12 = 0x78;
                goto LAB_1801517f4;
              }
/* novels.pipeline.short_fiction_parsers#32="```\\s*$" */
              local_70 = DAT_1803b6d30;
/* novels.pipeline.short_fiction_parsers#4="" */
              local_68 = DAT_1803b6c50;
              lVar4 = FUN_18014c500(param_1);
              if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
                if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                  (**(code **)(plVar5[1] + 0x30))(plVar5);
                }
                uVar12 = 0x78;
                plVar8 = local_res20;
              }
              else {
                plVar6 = (longlong *)FUN_18037f0c0();
                if (plVar6 == (longlong *)0x0) {
                  plVar8 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                    (**(code **)(plVar5[1] + 0x30))(plVar5);
                  }
                  uVar12 = 0x78;
                }
                else {
                  local_58 = local_70;
                  uVar12 = 0x78;
                  local_50 = local_68;
                  *(undefined4 *)(plVar3 + 5) = 0x78;
                  local_68 = plVar6;
                  local_48 = plVar7;
                  plVar13 = (longlong *)
/* novels.pipeline.short_fiction_parsers#31={"type": "tuple", "items": [{"tag": "a", "fileRange": [3999233, 3999240], "value": "flags"}]} */
                            FUN_18037da70(param_1,plVar5,&local_58,&local_68,DAT_1803b6d28);
                  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                    (**(code **)(plVar5[1] + 0x30))(plVar5);
                  }
                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                  }
                  if (plVar13 == (longlong *)0x0) {
LAB_1801517f4:
                    plVar8 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar13 = plVar7;
                  }
                  else {
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    lVar4 = FUN_18014c500(param_1);
                    if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                      FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
                      uVar12 = 0x79;
                      plVar8 = local_res20;
                    }
                    else {
                      plVar5 = (longlong *)FUN_18037f0c0();
                      plVar7 = plVar13;
                      if (plVar5 == (longlong *)0x0) {
                        uVar12 = 0x79;
                        goto LAB_1801517f4;
                      }
/* novels.pipeline.short_fiction_parsers#8="^\\s*===\\s*[A-Z0-9_ ]+\\s*===\\s*$" */
                      local_68 = DAT_1803b6c70;
/* novels.pipeline.short_fiction_parsers#4="" */
                      local_60 = DAT_1803b6c50;
                      lVar4 = FUN_18014c500(param_1);
                      if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                        FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
                        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                          (**(code **)(plVar5[1] + 0x30))(plVar5);
                        }
LAB_1801515a2:
                        uVar12 = 0x79;
                        plVar8 = local_res20;
                      }
                      else {
                        plVar6 = (longlong *)FUN_18037f0c0();
                        if (plVar6 == (longlong *)0x0) {
LAB_1801515c7:
                          plVar8 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                            (**(code **)(plVar5[1] + 0x30))(plVar5);
                          }
                        }
                        else {
                          lVar4 = FUN_18014c500(param_1);
                          if (lVar4 == 0) {
/* novels.pipeline.short_fiction_parsers#0="re" */
                            FUN_180373340(param_1,&local_res20,DAT_1803b6c30);
                            if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                            }
                            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                              uVar12 = 0x79;
                              plVar8 = local_res20;
                              goto LAB_180151804;
                            }
                            goto LAB_1801515a2;
                          }
                          plVar8 = (longlong *)FUN_18037f0c0();
                          if (plVar8 != (longlong *)0x0) {
                            plVar9 = (longlong *)FUN_180381230(plVar6,plVar8);
                            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                            }
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            if (plVar9 != (longlong *)0x0) {
                              local_58 = local_68;
                              uVar12 = 0x79;
                              local_50 = local_60;
                              *(undefined4 *)(plVar3 + 5) = 0x79;
                              local_res20 = plVar9;
                              local_48 = plVar13;
                              plVar6 = (longlong *)
                                       FUN_18037da70(param_1,plVar5,&local_58,&local_res20,
/* novels.pipeline.short_fiction_parsers#31={"type": "tuple", "items": [{"tag": "a", "fileRange": [3999233, 3999240], "value": "flags"}]} */
                                                     DAT_1803b6d28);
                              if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                (**(code **)(plVar5[1] + 0x30))(plVar5);
                              }
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              if (plVar6 != (longlong *)0x0) {
                                if ((-1 < (int)*plVar13) &&
                                   (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                                }
                                plVar13 = (longlong *)FUN_18037f0c0();
                                uVar12 = 0x7a;
                                plVar7 = plVar6;
                                if (plVar13 != (longlong *)0x0) {
                                  *(undefined4 *)(plVar3 + 5) = 0x7a;
                                  lVar4 = FUN_180376210(param_1,plVar13);
                                  if ((-1 < (int)*plVar13) &&
                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                                  }
                                  if (lVar4 != 0) {
                                    FUN_180001d80(param_1);
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                    }
                                    if ((int)*plVar1 < 0) {
                                      return lVar4;
                                    }
                                    *plVar1 = *plVar1 + -1;
                                    if (*plVar1 != 0) {
                                      return lVar4;
                                    }
                                    (**(code **)(plVar1[1] + 0x30))();
                                    return lVar4;
                                  }
                                }
                              }
                              goto LAB_1801517f4;
                            }
                            goto LAB_1801515c7;
                          }
                          plVar8 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                            (**(code **)(plVar5[1] + 0x30))(plVar5);
                          }
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                            uVar12 = 0x79;
                            goto LAB_180151804;
                          }
                        }
                        uVar12 = 0x79;
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
LAB_180151804:
  piVar2 = (int *)plVar8[5];
  if (piVar2 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar3,uVar12);
    plVar7 = (longlong *)plVar8[5];
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_18015185e:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1801518a7;
    pcVar10 = (code *)FUN_18038e2a0(plVar3,uVar12);
    *(int **)(pcVar10 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar7 = (longlong *)plVar8[5];
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_18015185e;
    pcVar11 = (code *)0x0;
  }
  plVar8[5] = (longlong)pcVar11;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1801518a7:
  FUN_18036d0a0(plVar3,&DAT_180391be4,plVar1,plVar13);
  plVar7 = DAT_1803b6bd0;
  if (plVar3 == DAT_1803b6bd0) {
    if ((-1 < (int)*DAT_1803b6bd0) && (*DAT_1803b6bd0 = *DAT_1803b6bd0 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b6bd0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))(plVar13);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
