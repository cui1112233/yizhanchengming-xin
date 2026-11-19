/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c6020 */

code * FUN_1802c6020(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  uint uVar1;
  int *piVar2;
  undefined8 *puVar3;
  longlong *plVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong lVar12;
  code *pcVar13;
  code *pcVar14;
  ulonglong uVar15;
  uint *puVar16;
  undefined8 uVar17;
  longlong *plVar18;
  longlong *plVar19;
  bool bVar20;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_98;
  longlong local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  longlong *local_40;
  
  plVar4 = DAT_1803c4c00;
  plVar19 = (longlong *)0x0;
  plVar8 = (longlong *)*param_3;
  plVar10 = (longlong *)0x0;
  plVar11 = (longlong *)param_3[1];
  plVar9 = (longlong *)0x0;
  plVar18 = (longlong *)0x0;
  local_80 = (longlong *)0x0;
  local_res20 = (longlong *)0x0;
  local_70 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_98 = (longlong *)0x0;
  local_88 = DAT_180391e98;
  local_78 = plVar8;
  local_68 = plVar11;
  if (DAT_1803c4c00 == (longlong *)0x0) {
LAB_1802c60b8:
    DAT_1803c4c00 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4bd0,DAT_1803ca468,0x38);
  }
  else if ((1 < *DAT_1803c4c00) || (DAT_1803c4c00[2] != 0)) {
    if ((DAT_1803c4c00 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4c00 && (*DAT_1803c4c00 = *DAT_1803c4c00 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1802c60b8;
  }
  plVar4 = DAT_1803c4c00;
  FUN_180001d00(param_1,DAT_1803c4c00);
  lVar6 = FUN_1802c2140(param_1);
  if (lVar6 == 0) {
    FUN_180373340(param_1,&local_88,DAT_1803c4d30);
    uVar17 = 0x59;
    lVar6 = local_88;
    local_98 = plVar10;
    local_res20 = plVar19;
  }
  else {
    uVar17 = 0x59;
    *(undefined4 *)(plVar4 + 5) = 0x59;
    local_50 = (longlong *)FUN_180376770(param_1,lVar6,plVar11);
    if (local_50 == (longlong *)0x0) {
      lVar6 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      local_98 = plVar10;
      local_res20 = plVar19;
    }
    else {
      local_80 = local_50;
      lVar6 = FUN_180373860(local_50);
      plVar11 = (longlong *)(param_1 + 0x60);
      if (((lVar6 < 0) && (*plVar11 != 0)) ||
         (plVar7 = (longlong *)PyLong_FromSsize_t(lVar6), plVar7 == (longlong *)0x0)) {
        lVar6 = *plVar11;
        *plVar11 = 0;
        uVar17 = 0x5a;
        local_98 = plVar10;
        local_res20 = plVar19;
      }
      else {
        uVar15 = (ulonglong)plVar7[2] >> 3;
        lVar6 = (1 - (ulonglong)((uint)plVar7[2] & 3)) * uVar15;
        if (lVar6 == 1) {
          bVar20 = false;
          lVar6 = uVar15 - 1;
          if (-1 < lVar6) {
            puVar16 = (uint *)((longlong)plVar7 + (uVar15 + 5) * 4);
            do {
              uVar1 = *puVar16;
              if (uVar1 != 10) {
                bVar20 = uVar1 < 10;
                if ((*(byte *)(plVar7 + 2) & 2) != 0) {
                  bVar20 = 9 < uVar1;
                }
                break;
              }
              puVar16 = puVar16 + -1;
              lVar6 = lVar6 + -1;
            } while (-1 < lVar6);
          }
        }
        else {
          bVar20 = SUB81((ulonglong)(lVar6 + -1) >> 0x3f,0);
        }
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (bVar20 != false) goto LAB_1802c64fe;
        lVar6 = FUN_1802c2140(param_1);
        if (lVar6 == 0) {
          FUN_180373340(param_1,&local_88,DAT_1803c4d30);
          uVar17 = 0x5c;
          lVar6 = local_88;
          local_98 = (longlong *)0x0;
          plVar9 = plVar18;
          local_res20 = (longlong *)0x0;
        }
        else {
          uVar17 = 0x5c;
          *(undefined4 *)(plVar4 + 5) = 0x5c;
          local_res20 = (longlong *)FUN_180376770(param_1,lVar6,plVar8);
          if (local_res20 == (longlong *)0x0) {
            lVar6 = *(longlong *)(param_1 + 0x60);
            local_98 = (longlong *)0x0;
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar9 = plVar18;
            local_res20 = (longlong *)0x0;
          }
          else {
            plVar8 = (longlong *)FUN_18037f0c0();
            uVar17 = 0x5d;
            if (plVar8 == (longlong *)0x0) {
              lVar6 = *(longlong *)(param_1 + 0x60);
              local_98 = (longlong *)0x0;
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar9 = plVar18;
            }
            else {
              *(undefined4 *)(plVar4 + 5) = 0x5d;
              local_60 = (longlong *)FUN_180376770(param_1,plVar8,local_80);
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if (local_60 == (longlong *)0x0) {
                lVar6 = *(longlong *)(param_1 + 0x60);
                local_98 = (longlong *)0x0;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar9 = plVar18;
              }
              else {
                local_70 = local_60;
                iVar5 = FUN_180383a30(local_60,DAT_1803ab260);
                if (iVar5 == -1) {
                  lVar6 = *(longlong *)(param_1 + 0x60);
                  uVar17 = 0x5e;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  local_98 = (longlong *)0x0;
                  plVar9 = plVar18;
                }
                else {
                  if (iVar5 == 1) goto LAB_1802c64fe;
                  plVar8 = (longlong *)FUN_18037f0c0();
                  if (plVar8 == (longlong *)0x0) {
                    lVar6 = *(longlong *)(param_1 + 0x60);
                    uVar17 = 0x60;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    local_98 = (longlong *)0x0;
                    plVar9 = plVar18;
                  }
                  else {
                    lVar6 = FUN_180373860(local_80);
                    plVar11 = (longlong *)(param_1 + 0x60);
                    if (((lVar6 < 0) && (*plVar11 != 0)) ||
                       (plVar9 = (longlong *)PyLong_FromSsize_t(lVar6), plVar9 == (longlong *)0x0))
                    {
                      lVar6 = *plVar11;
                      *plVar11 = 0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                      }
                      uVar17 = 0x60;
                      local_98 = (longlong *)0x0;
                      plVar9 = plVar18;
                    }
                    else {
                      plVar10 = (longlong *)FUN_180380b80(local_60,plVar9);
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      plVar9 = local_98;
                      if (plVar10 == (longlong *)0x0) {
                        lVar6 = *plVar11;
                        *plVar11 = 0;
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        uVar17 = 0x60;
                      }
                      else {
                        uVar17 = 0x60;
                        local_48 = local_80;
                        *(undefined4 *)(plVar4 + 5) = 0x60;
                        local_40 = plVar10;
                        plVar11 = (longlong *)FUN_180377750(param_1,plVar8,&local_48);
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                        }
                        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                        if (plVar11 == (longlong *)0x0) {
                          lVar6 = *(longlong *)(param_1 + 0x60);
                          local_98 = (longlong *)0x0;
                          *(longlong *)(param_1 + 0x60) = 0;
                          plVar9 = local_98;
                        }
                        else {
                          iVar5 = FUN_180386510(plVar11,DAT_1803ab260);
                          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                          }
                          plVar8 = (longlong *)(param_1 + 0x60);
                          if (iVar5 == -1) {
                            local_98 = (longlong *)0x0;
                            lVar6 = *(longlong *)(param_1 + 0x60);
                            *(longlong *)(param_1 + 0x60) = 0;
                            plVar9 = local_98;
                          }
                          else {
                            if (iVar5 == 1) {
                              plVar9 = (longlong *)0x0;
                              goto LAB_1802c64fe;
                            }
                            lVar6 = FUN_1802c2030(param_1);
                            plVar11 = local_60;
                            if (lVar6 == 0) {
                              FUN_180373340(param_1,&local_88,DAT_1803c4d38);
                              local_98 = (longlong *)0x0;
                              uVar17 = 0x62;
                              lVar6 = local_88;
                              plVar9 = local_98;
                            }
                            else {
                              uVar17 = 0x62;
                              local_48 = local_78;
                              *(undefined4 *)(plVar4 + 5) = 0x62;
                              local_40 = local_60;
                              local_res18 = (longlong *)FUN_180377750(param_1,lVar6,&local_48);
                              if (local_res18 == (longlong *)0x0) {
                                lVar6 = *plVar8;
                                local_98 = (longlong *)0x0;
                                *plVar8 = 0;
                                plVar9 = local_98;
                              }
                              else {
                                lVar6 = FUN_1802c2030(param_1);
                                if (lVar6 == 0) {
                                  FUN_180373340(param_1,&local_88,DAT_1803c4d38);
                                  uVar17 = 99;
                                  plVar9 = (longlong *)0x0;
                                  lVar6 = local_88;
                                  local_98 = local_res18;
                                }
                                else {
                                  lVar12 = FUN_180373860(local_80);
                                  if (((lVar12 < 0) && (*plVar8 != 0)) ||
                                     (plVar8 = (longlong *)PyLong_FromSsize_t(lVar12),
                                     plVar8 == (longlong *)0x0)) {
                                    uVar17 = 99;
                                  }
                                  else {
                                    plVar11 = (longlong *)FUN_180380b80(plVar11,plVar8);
                                    if ((-1 < (int)*plVar8) &&
                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                    }
                                    uVar17 = 99;
                                    if (plVar11 == (longlong *)0x0) {
                                      plVar9 = (longlong *)0x0;
                                    }
                                    else {
                                      local_60 = local_78;
                                      *(undefined4 *)(plVar4 + 5) = 99;
                                      local_58 = plVar11;
                                      plVar9 = (longlong *)FUN_180377750(param_1,lVar6,&local_60);
                                      if ((-1 < (int)*plVar11) &&
                                         (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                                      }
                                      if (plVar9 == (longlong *)0x0) {
                                        plVar9 = (longlong *)0x0;
                                      }
                                      else {
                                        iVar5 = FUN_180383a30(local_res18,DAT_1803ab260);
                                        if (iVar5 != -1) {
                                          if (iVar5 != 1) {
                                            iVar5 = FUN_180383a30(plVar9,DAT_1803ab260);
                                            if (iVar5 == -1) goto LAB_1802c67e1;
                                            if (iVar5 != 1) {
                                              FUN_180001d80(param_1);
                                              lVar6 = *(longlong *)(param_1 + 0x10);
                                              pcVar13 = *(code **)(lVar6 + 0x41da0);
                                              if (pcVar13 == (code *)0x0) {
                                                pcVar13 = (code *)FUN_18035ebd0(PyTuple_Type_exref,2
                                                                               );
                                              }
                                              else {
                                                *(undefined8 *)(lVar6 + 0x41da0) =
                                                     *(undefined8 *)(pcVar13 + 0x18);
                                                *(int *)(lVar6 + 0x41e3c) =
                                                     *(int *)(lVar6 + 0x41e3c) + -1;
                                                *(undefined8 *)pcVar13 = 1;
                                              }
                                              pcVar14 = pcVar13 + -0x10;
                                              *(undefined8 *)(pcVar13 + 0x18) = 0;
                                              *(undefined8 *)(pcVar13 + 0x20) = 0;
                                              lVar6 = _PyThreadState_GetCurrent();
                                              lVar6 = *(longlong *)
                                                       (*(longlong *)(lVar6 + 0x10) + 0xd0);
                                              puVar3 = *(undefined8 **)(lVar6 + 8);
                                              *puVar3 = pcVar14;
                                              *(ulonglong *)(pcVar13 + -8) =
                                                   *(ulonglong *)(pcVar13 + -8) & 3;
                                              *(ulonglong *)(pcVar13 + -8) =
                                                   *(ulonglong *)(pcVar13 + -8) | (ulonglong)puVar3;
                                              *(longlong *)pcVar14 = lVar6;
                                              *(code **)(lVar6 + 8) = pcVar14;
                                              iVar5 = (int)*local_res18 + 1;
                                              if (iVar5 != 0) {
                                                *(int *)local_res18 = iVar5;
                                              }
                                              *(longlong **)(pcVar13 + 0x18) = local_res18;
                                              iVar5 = (int)*plVar9 + 1;
                                              if (iVar5 != 0) {
                                                *(int *)plVar9 = iVar5;
                                              }
                                              *(longlong **)(pcVar13 + 0x20) = plVar9;
                                              goto LAB_1802c651e;
                                            }
                                          }
LAB_1802c64fe:
                                          pcVar13 = _Py_NoneStruct_exref;
                                          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                                            *(int *)_Py_NoneStruct_exref =
                                                 *(int *)_Py_NoneStruct_exref + 1;
                                          }
                                          FUN_180001d80(param_1);
LAB_1802c651e:
                                          if ((-1 < (int)*local_50) &&
                                             (*local_50 = *local_50 + -1, *local_50 == 0)) {
                                            (**(code **)(local_80[1] + 0x30))();
                                          }
                                          if (((local_res20 != (longlong *)0x0) &&
                                              (-1 < (int)*local_res20)) &&
                                             (*local_res20 = *local_res20 + -1, *local_res20 == 0))
                                          {
                                            (**(code **)(local_res20[1] + 0x30))();
                                          }
                                          if (((local_70 != (longlong *)0x0) &&
                                              (-1 < (int)*local_70)) &&
                                             (*local_70 = *local_70 + -1, *local_70 == 0)) {
                                            (**(code **)(local_70[1] + 0x30))();
                                          }
                                          if (((local_res18 != (longlong *)0x0) &&
                                              (-1 < (int)*local_res18)) &&
                                             (*local_res18 = *local_res18 + -1, *local_res18 == 0))
                                          {
                                            (**(code **)(local_res18[1] + 0x30))(local_res18);
                                          }
                                          if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9))
                                             && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                                          }
                                          if ((-1 < (int)*local_78) &&
                                             (*local_78 = *local_78 + -1, *local_78 == 0)) {
                                            (**(code **)(local_78[1] + 0x30))();
                                          }
                                          if ((int)*local_68 < 0) {
                                            return pcVar13;
                                          }
                                          *local_68 = *local_68 + -1;
                                          if (*local_68 != 0) {
                                            return pcVar13;
                                          }
                                          (**(code **)(local_68[1] + 0x30))(local_68);
                                          return pcVar13;
                                        }
LAB_1802c67e1:
                                        uVar17 = 100;
                                      }
                                    }
                                  }
                                  lVar6 = *(longlong *)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  local_98 = local_res18;
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
  piVar2 = *(int **)(lVar6 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar13 = (code *)FUN_18038e2a0(plVar4,uVar17);
    plVar8 = *(longlong **)(lVar6 + 0x28);
    pcVar14 = pcVar13;
    if (pcVar13 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
LAB_1802c6857:
      pcVar13 = pcVar14;
      if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar4) goto LAB_1802c68a0;
    pcVar13 = (code *)FUN_18038e2a0(plVar4,uVar17);
    *(int **)(pcVar13 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar8 = *(longlong **)(lVar6 + 0x28);
    pcVar14 = pcVar13;
    if (pcVar13 != _Py_NoneStruct_exref) goto LAB_1802c6857;
    pcVar14 = (code *)0x0;
  }
  *(code **)(lVar6 + 0x28) = pcVar14;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
LAB_1802c68a0:
  plVar10 = local_70;
  plVar11 = local_78;
  FUN_18036d0a0(plVar4,"ooooooo",local_78,local_68,local_80,local_res20,local_70,local_98,plVar9);
  plVar8 = DAT_1803c4c00;
  if (plVar4 == DAT_1803c4c00) {
    if ((-1 < (int)*DAT_1803c4c00) && (*DAT_1803c4c00 = *DAT_1803c4c00 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803c4c00 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
     (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if (((local_res20 != (longlong *)0x0) && (-1 < (int)*local_res20)) &&
     (*local_res20 = *local_res20 + -1, *local_res20 == 0)) {
    (**(code **)(local_res20[1] + 0x30))(local_res20);
  }
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))(local_98);
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*local_68) && (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
  }
  plVar8 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar6;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  return (code *)0x0;
}

