/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180150000 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

int * FUN_180150000(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  uint uVar1;
  undefined8 *puVar2;
  ulonglong uVar3;
  longlong lVar4;
  undefined8 uVar5;
  int iVar6;
  longlong lVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  longlong *plVar13;
  longlong *plVar14;
  ulonglong uVar15;
  longlong *plVar16;
  int *piVar17;
  uint *puVar18;
  undefined8 uVar19;
  longlong *plVar20;
  bool bVar21;
  undefined1 auStack_148 [32];
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  undefined8 local_60;
  longlong *local_58;
  longlong *local_50;
  ulonglong local_48;

  plVar13 = DAT_1803b6a88;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_148;
  local_b8 = (longlong *)*param_3;
  plVar9 = (longlong *)0x0;
  local_90 = (longlong *)param_3[1];
  plVar20 = (longlong *)0x0;
  local_c0 = DAT_180391d20;
  local_a8 = (longlong *)0x0;
  local_a0 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_88 = (longlong *)0x0;
  local_98 = (longlong *)0x0;
  local_e0 = (longlong *)0x0;
  local_c8 = (longlong *)0x0;
  local_d0 = DAT_1803b6a88;
  if (DAT_1803b6a88 == (longlong *)0x0) {
LAB_1801500a5:
    DAT_1803b6a88 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6ac0,DAT_1803ca758,0x48);
  }
  else if ((1 < *DAT_1803b6a88) || (DAT_1803b6a88[2] != 0)) {
    if ((DAT_1803b6a88 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6a88 && (*DAT_1803b6a88 = *DAT_1803b6a88 + -1, *plVar13 == 0)))) {
      (**(code **)(plVar13[1] + 0x30))(plVar13);
    }
    goto LAB_1801500a5;
  }
  plVar13 = DAT_1803b6a88;
  local_d0 = DAT_1803b6a88;
  FUN_180001d00(param_1,DAT_1803b6a88);
  lVar7 = FUN_18014c500(param_1);
  if (lVar7 == 0) {
    FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
    uVar19 = 0x67;
    lVar7 = local_c0;
    plVar16 = plVar9;
  }
  else {
    plVar8 = (longlong *)FUN_18037f0c0();
    piVar17 = DAT_1803b6d00;
    if (plVar8 == (longlong *)0x0) {
      lVar7 = *(longlong *)(param_1 + 0x60);
      uVar19 = 0x67;
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar9 = plVar20;
      plVar16 = (longlong *)0x0;
    }
    else {
      lVar7 = *(longlong *)(param_1 + 0x10);
      plVar10 = *(longlong **)(lVar7 + 0x41da8);
      if (plVar10 == (longlong *)0x0) {
        plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
      }
      else {
        *(longlong *)(lVar7 + 0x41da8) = plVar10[3];
        *(int *)(lVar7 + 0x41e40) = *(int *)(lVar7 + 0x41e40) + -1;
        *plVar10 = 1;
      }
      plVar10[3] = 0;
      plVar10[4] = 0;
      plVar10[5] = 0;
      plVar16 = plVar10 + -2;
      lVar7 = _PyThreadState_GetCurrent();
      lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
      puVar2 = *(undefined8 **)(lVar7 + 8);
      *puVar2 = plVar16;
      plVar10[-1] = plVar10[-1] & 3;
      plVar10[-1] = plVar10[-1] | (ulonglong)puVar2;
      *plVar16 = lVar7;
      *(longlong **)(lVar7 + 8) = plVar16;
      if (*piVar17 + 1 != 0) {
        *piVar17 = *piVar17 + 1;
      }
      plVar10[3] = (longlong)piVar17;
      lVar7 = FUN_180375520(param_1,local_90,DAT_1803b6c50);
      plVar16 = plVar20;
      if (lVar7 == 0) {
        lVar7 = *(longlong *)(param_1 + 0x60);
        uVar19 = 0x67;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        plVar16 = plVar9;
        if ((-1 < (int)*plVar10) &&
           (*plVar10 = *plVar10 + -1, plVar9 = plVar20, plVar16 = plVar20, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
      }
      else {
        plVar10[4] = lVar7;
        piVar17 = DAT_1803b6d08;
        if (*DAT_1803b6d08 + 1 != 0) {
          *DAT_1803b6d08 = *DAT_1803b6d08 + 1;
        }
        plVar10[5] = (longlong)piVar17;
        plVar9 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar10);
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (plVar9 == (longlong *)0x0) {
          lVar7 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          uVar19 = 0x67;
          plVar9 = plVar20;
        }
        else {
          uVar19 = 0x67;
          *(undefined4 *)(plVar13 + 5) = 0x67;
          plVar10 = (longlong *)FUN_180376770(param_1,plVar8,plVar9);
          local_d8 = plVar10;
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (plVar10 == (longlong *)0x0) {
            lVar7 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar9 = plVar20;
          }
          else {
            local_a8 = plVar10;
            lVar7 = FUN_18014c500(param_1);
            if (lVar7 == 0) {
              FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
              uVar19 = 0x68;
              lVar7 = local_c0;
              plVar9 = plVar20;
            }
            else {
              plVar9 = (longlong *)FUN_18037f0c0();
              piVar17 = DAT_1803b6c48;
              if (plVar9 == (longlong *)0x0) {
                lVar7 = *(longlong *)(param_1 + 0x60);
                uVar19 = 0x68;
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar9 = plVar20;
              }
              else {
                lVar7 = *(longlong *)(param_1 + 0x10);
                plVar13 = *(longlong **)(lVar7 + 0x41da8);
                if (plVar13 == (longlong *)0x0) {
                  plVar13 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3);
                }
                else {
                  *(longlong *)(lVar7 + 0x41da8) = plVar13[3];
                  *(int *)(lVar7 + 0x41e40) = *(int *)(lVar7 + 0x41e40) + -1;
                  *plVar13 = 1;
                }
                plVar8 = plVar13 + -2;
                plVar13[3] = 0;
                plVar13[4] = 0;
                plVar13[5] = 0;
                lVar7 = _PyThreadState_GetCurrent();
                lVar7 = *(longlong *)(*(longlong *)(lVar7 + 0x10) + 0xd0);
                puVar2 = *(undefined8 **)(lVar7 + 8);
                *puVar2 = plVar8;
                plVar13[-1] = plVar13[-1] & 3;
                plVar13[-1] = plVar13[-1] | (ulonglong)puVar2;
                *plVar8 = lVar7;
                *(longlong **)(lVar7 + 8) = plVar8;
                if (*piVar17 + 1 != 0) {
                  *piVar17 = *piVar17 + 1;
                }
                plVar13[3] = (longlong)piVar17;
                lVar7 = FUN_180375520(param_1,plVar10,DAT_1803b6c50);
                if (lVar7 == 0) {
                  lVar7 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                  }
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
                }
                else {
                  plVar13[4] = lVar7;
                  piVar17 = DAT_1803b6c58;
                  if (*DAT_1803b6c58 + 1 != 0) {
                    *DAT_1803b6c58 = *DAT_1803b6c58 + 1;
                  }
                  plVar13[5] = (longlong)piVar17;
                  plVar8 = (longlong *)PyUnicode_Join(DAT_1803b6c50,plVar13);
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
                  if (plVar8 == (longlong *)0x0) {
                    lVar7 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                      pcVar11 = *(code **)(plVar9[1] + 0x30);
                      plVar13 = plVar9;
LAB_1801504bc:
                      (*pcVar11)(plVar13);
                    }
                  }
                  else {
                    lVar7 = FUN_18014c500(param_1);
                    if (lVar7 == 0) {
                      FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      lVar7 = local_c0;
                      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                        lVar7 = local_c0;
                      }
                    }
                    else {
                      plVar13 = (longlong *)FUN_18037f0c0();
                      if (plVar13 == (longlong *)0x0) {
LAB_18015056d:
                        lVar7 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                          pcVar11 = *(code **)(plVar8[1] + 0x30);
                          plVar13 = plVar8;
                          goto LAB_1801504bc;
                        }
                      }
                      else {
                        lVar7 = FUN_18014c500(param_1);
                        if (lVar7 == 0) {
                          FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                          }
                          lVar7 = local_c0;
                          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                            (**(code **)(plVar13[1] + 0x30))(plVar13);
                            lVar7 = local_c0;
                          }
                        }
                        else {
                          plVar10 = (longlong *)FUN_18037f0c0();
                          if (plVar10 != (longlong *)0x0) {
                            plVar14 = (longlong *)FUN_180381230(plVar13,plVar10);
                            if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                              (**(code **)(plVar13[1] + 0x30))(plVar13);
                            }
                            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                            if (plVar14 == (longlong *)0x0) goto LAB_18015056d;
                            uVar19 = 0x68;
                            *(undefined4 *)(local_d0 + 5) = 0x68;
                            local_70 = plVar8;
                            local_68 = plVar14;
                            plVar20 = (longlong *)FUN_180377750(param_1,plVar9,&local_70);
                            local_80 = plVar20;
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                              (**(code **)(plVar14[1] + 0x30))(plVar14);
                            }
                            plVar8 = local_b8;
                            plVar13 = local_d0;
                            if (plVar20 != (longlong *)0x0) {
                              uVar19 = 0x69;
                              *(undefined4 *)(local_d0 + 5) = 0x69;
                              local_a0 = plVar20;
                              plVar9 = (longlong *)
                                       FUN_18037e6c0(param_1,plVar20,DAT_1803b6c78,local_b8);
                              if (plVar9 != (longlong *)0x0) {
                                plVar10 = (longlong *)FUN_180372690(param_1,plVar9);
                                local_78 = plVar10;
                                if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                }
                                if (plVar10 != (longlong *)0x0) {
                                  local_b0 = plVar10;
                                  lVar7 = FUN_180373860(plVar10);
                                  if ((lVar7 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) {
                                    plVar9 = (longlong *)0x0;
                                  }
                                  else {
                                    plVar9 = (longlong *)PyLong_FromSsize_t(lVar7);
                                  }
                                  uVar3 = plVar9[2];
                                  uVar15 = uVar3 >> 3;
                                  lVar7 = (1 - (ulonglong)((uint)uVar3 & 3)) * uVar15;
                                  if (lVar7 == 1) {
                                    bVar21 = false;
                                    lVar7 = uVar15 - 1;
                                    if (-1 < lVar7) {
                                      puVar18 = (uint *)((longlong)plVar9 + (uVar15 + 5) * 4);
                                      do {
                                        uVar1 = *puVar18;
                                        if (uVar1 != 2) {
                                          bVar21 = uVar1 < 2;
                                          if ((uVar3 & 2) != 0) {
                                            bVar21 = 1 < uVar1;
                                          }
                                          break;
                                        }
                                        puVar18 = puVar18 + -1;
                                        lVar7 = lVar7 + -1;
                                      } while (-1 < lVar7);
                                    }
                                  }
                                  else {
                                    bVar21 = SUB81((ulonglong)(lVar7 + -1) >> 0x3f,0);
                                  }
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
                                  piVar17 = DAT_1803b6c50;
                                  if (bVar21 != false) {
                                    if (*DAT_1803b6c50 + 1 != 0) {
                                      *DAT_1803b6c50 = *DAT_1803b6c50 + 1;
                                    }
                                    goto LAB_180150da6;
                                  }
                                  plVar9 = (longlong *)
                                           FUN_18014b9b0(param_1,plVar10,DAT_1803ab258,1);
                                  if (plVar9 == (longlong *)0x0) {
                                    uVar19 = 0x6c;
                                  }
                                  else {
                                    uVar19 = 0x6d;
                                    *(undefined4 *)(plVar13 + 5) = 0x6d;
                                    local_88 = plVar9;
                                    plVar9 = (longlong *)FUN_18037e560(param_1,plVar9,DAT_1803b6c80)
                                    ;
                                    if (plVar9 != (longlong *)0x0) {
                                      local_98 = plVar9;
                                      plVar9 = (longlong *)
                                               FUN_18014b800(param_1,plVar9,_Py_NoneStruct_exref);
                                      plVar20 = (longlong *)FUN_180024e50(param_1,plVar8,plVar9);
                                      if ((-1 < (int)*plVar9) &&
                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                                      }
                                      if (plVar20 == (longlong *)0x0) {
                                        uVar19 = 0x6e;
                                      }
                                      else {
                                        plVar9 = (longlong *)FUN_18037f0c0();
                                        if ((-1 < (int)*plVar20) &&
                                           (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                                          (**(code **)(plVar20[1] + 0x30))(plVar20);
                                        }
                                        uVar19 = 0x6e;
                                        if (plVar9 != (longlong *)0x0) {
                                          *(undefined4 *)(plVar13 + 5) = 0x6e;
                                          plVar16 = (longlong *)
                                                    FUN_1803770d0(param_1,plVar9,DAT_1803b6c90);
                                          if ((-1 < (int)*plVar9) &&
                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                                          }
                                          if (plVar16 != (longlong *)0x0) {
                                            local_e0 = plVar16;
                                            lVar7 = FUN_18014c500(param_1);
                                            if (lVar7 == 0) {
                                              FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
                                              uVar19 = 0x6f;
                                              lVar7 = local_c0;
                                              plVar9 = local_c8;
                                            }
                                            else {
                                              plVar20 = (longlong *)FUN_18037f0c0();
                                              uVar5 = DAT_1803b6d10;
                                              if (plVar20 == (longlong *)0x0) {
                                                lVar7 = *(longlong *)(param_1 + 0x60);
                                                uVar19 = 0x6f;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                plVar9 = local_c8;
                                              }
                                              else {
                                                lVar7 = FUN_18014c500(param_1);
                                                if (lVar7 == 0) {
                                                  FUN_180373340(param_1,&local_c0,DAT_1803b6c30);
                                                  if ((-1 < (int)*plVar20) &&
                                                     (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                                                    (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                  }
LAB_180150a54:
                                                  uVar19 = 0x71;
                                                  lVar7 = local_c0;
                                                  plVar9 = local_c8;
                                                }
                                                else {
                                                  plVar9 = (longlong *)FUN_18037f0c0();
                                                  if (plVar9 == (longlong *)0x0) {
                                                    lVar7 = *(longlong *)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    if ((-1 < (int)*plVar20) &&
                                                       (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                                                      (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                    }
                                                    uVar19 = 0x71;
                                                    plVar9 = local_c8;
                                                  }
                                                  else {
                                                    lVar7 = FUN_18014c500(param_1);
                                                    if (lVar7 == 0) {
                                                      FUN_180373340(param_1,&local_c0,DAT_1803b6c30)
                                                      ;
                                                      if ((-1 < (int)*plVar20) &&
                                                         (*plVar20 = *plVar20 + -1, *plVar20 == 0))
                                                      {
                                                        (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                      }
                                                      if (((int)*plVar9 < 0) ||
                                                         (*plVar9 = *plVar9 + -1, *plVar9 != 0))
                                                      goto LAB_180150a54;
                                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                      uVar19 = 0x71;
                                                      lVar7 = local_c0;
                                                      plVar9 = local_c8;
                                                    }
                                                    else {
                                                      plVar8 = (longlong *)FUN_18037f0c0();
                                                      if (plVar8 == (longlong *)0x0) {
                                                        lVar7 = *(longlong *)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        if ((-1 < (int)*plVar20) &&
                                                           (*plVar20 = *plVar20 + -1, *plVar20 == 0)
                                                           ) {
                                                          (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                        }
                                                        uVar19 = 0x71;
                                                        plVar9 = local_c8;
                                                        plVar16 = local_e0;
                                                      }
                                                      else {
                                                        plVar13 = (longlong *)
                                                                  FUN_180381230(plVar9,plVar8);
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                        }
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                        }
                                                        if (plVar13 == (longlong *)0x0) {
                                                          lVar7 = *(longlong *)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          if ((-1 < (int)*plVar20) &&
                                                             (*plVar20 = *plVar20 + -1,
                                                             *plVar20 == 0)) {
                                                            (**(code **)(plVar20[1] + 0x30))
                                                                      (plVar20);
                                                          }
                                                          uVar19 = 0x71;
                                                          plVar13 = local_d0;
                                                          plVar9 = local_c8;
                                                          plVar16 = local_e0;
                                                        }
                                                        else {
                                                          uVar19 = 0x6f;
                                                          local_60 = uVar5;
                                                          *(undefined4 *)(local_d0 + 5) = 0x6f;
                                                          local_58 = plVar16;
                                                          local_50 = plVar13;
                                                          plVar9 = (longlong *)
                                                                   FUN_180378890(param_1,plVar20,
                                                                                 &local_60);
                                                          if ((-1 < (int)*plVar20) &&
                                                             (*plVar20 = *plVar20 + -1,
                                                             *plVar20 == 0)) {
                                                            (**(code **)(plVar20[1] + 0x30))
                                                                      (plVar20);
                                                          }
                                                          if ((-1 < (int)*plVar13) &&
                                                             (*plVar13 = *plVar13 + -1,
                                                             *plVar13 == 0)) {
                                                            (**(code **)(plVar13[1] + 0x30))
                                                                      (plVar13);
                                                          }
                                                          if (plVar9 == (longlong *)0x0) {
                                                            lVar7 = *(longlong *)(param_1 + 0x60);
                                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                                            plVar13 = local_d0;
                                                            plVar9 = local_c8;
                                                            plVar16 = local_e0;
                                                          }
                                                          else {
                                                            local_c8 = plVar9;
                                                            iVar6 = FUN_180001510(plVar9);
                                                            plVar13 = local_d0;
                                                            uVar19 = 0x73;
                                                            if (iVar6 != -1) {
                                                              if (iVar6 == 0) {
                                                                iVar6 = (int)*plVar16 + 1;
                                                                if (iVar6 != 0) {
                                                                  *(int *)plVar16 = iVar6;
                                                                }
LAB_180150d2f:
                                                                plVar20 = (longlong *)
                                                                          FUN_18037f0c0();
                                                                if ((-1 < (int)*plVar16) &&
                                                                   (*plVar16 = *plVar16 + -1,
                                                                   *plVar16 == 0)) {
                                                                  (**(code **)(plVar16[1] + 0x30))
                                                                            (plVar16);
                                                                }
                                                                if (plVar20 != (longlong *)0x0) {
                                                                  *(undefined4 *)(plVar13 + 5) =
                                                                       0x73;
                                                                  piVar17 = (int *)FUN_180376210(
                                                  param_1,plVar20);
                                                  if ((-1 < (int)*plVar20) &&
                                                     (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                                                    (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                  }
                                                  plVar20 = local_80;
                                                  plVar10 = local_78;
                                                  plVar8 = local_b8;
                                                  if (piVar17 != (int *)0x0) {
LAB_180150da6:
                                                    FUN_180001d80(param_1);
                                                    if ((-1 < (int)*local_d8) &&
                                                       (*local_d8 = *local_d8 + -1, *local_d8 == 0))
                                                    {
                                                      (**(code **)(local_a8[1] + 0x30))();
                                                    }
                                                    if ((-1 < (int)*plVar20) &&
                                                       (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
                                                      (**(code **)(local_a0[1] + 0x30))();
                                                    }
                                                    if ((-1 < (int)*plVar10) &&
                                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                      (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                    }
                                                    if (((local_88 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_88)) &&
                                                       (*local_88 = *local_88 + -1, *local_88 == 0))
                                                    {
                                                      (**(code **)(local_88[1] + 0x30))();
                                                    }
                                                    if (((local_98 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_98)) &&
                                                       (*local_98 = *local_98 + -1, *local_98 == 0))
                                                    {
                                                      (**(code **)(local_98[1] + 0x30))();
                                                    }
                                                    if (((local_e0 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_e0)) &&
                                                       (*local_e0 = *local_e0 + -1, *local_e0 == 0))
                                                    {
                                                      (**(code **)(local_e0[1] + 0x30))();
                                                    }
                                                    if (((local_c8 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_c8)) &&
                                                       (*local_c8 = *local_c8 + -1, *local_c8 == 0))
                                                    {
                                                      (**(code **)(local_c8[1] + 0x30))();
                                                    }
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                    }
                                                    if ((int)*local_90 < 0) {
                                                      return piVar17;
                                                    }
                                                    *local_90 = *local_90 + -1;
                                                    if (*local_90 != 0) {
                                                      return piVar17;
                                                    }
                                                    (**(code **)(local_90[1] + 0x30))();
                                                    return piVar17;
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    *(undefined4 *)(local_d0 + 5) = 0x73;
                                                    plVar20 = (longlong *)
                                                              FUN_18037e560(param_1,plVar9,
                                                                            DAT_1803b6cb0);
                                                    if (plVar20 != (longlong *)0x0) {
                                                      plVar8 = (longlong *)
                                                               FUN_18014b8e0(param_1,plVar20);
                                                      if ((-1 < (int)*plVar20) &&
                                                         (*plVar20 = *plVar20 + -1, *plVar20 == 0))
                                                      {
                                                        (**(code **)(plVar20[1] + 0x30))(plVar20);
                                                      }
                                                      plVar16 = (longlong *)
                                                                FUN_180024e50(param_1,plVar16,plVar8
                                                                             );
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                      }
                                                      if (plVar16 != (longlong *)0x0)
                                                      goto LAB_180150d2f;
                                                    }
                                                  }
                                                  }
                                                  lVar7 = *(longlong *)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  plVar16 = local_e0;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                }
                                              }
                                            }
                                            goto LAB_1801504cb;
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                            lVar7 = *(longlong *)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            plVar9 = local_c8;
                            plVar16 = local_c8;
                            goto LAB_1801504cb;
                          }
                          lVar7 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                          }
                          if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                            pcVar11 = *(code **)(plVar13[1] + 0x30);
                            goto LAB_1801504bc;
                          }
                        }
                      }
                    }
                  }
                }
                uVar19 = 0x68;
                plVar13 = local_d0;
                plVar9 = plVar20;
              }
            }
          }
        }
      }
    }
  }
LAB_1801504cb:
  piVar17 = *(int **)(lVar7 + 0x28);
  if (piVar17 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar13,uVar19);
    plVar20 = *(longlong **)(lVar7 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
LAB_180150f0f:
      pcVar11 = pcVar12;
      if ((pcVar12 != (code *)0x0) && (*(int *)pcVar12 + 1 != 0)) {
        *(int *)pcVar12 = *(int *)pcVar12 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar17 + 6) == plVar13) goto LAB_180150f58;
    pcVar11 = (code *)FUN_18038e2a0(plVar13,uVar19);
    *(int **)(pcVar11 + 0x10) = piVar17;
    if (*piVar17 + 1 != 0) {
      *piVar17 = *piVar17 + 1;
    }
    plVar20 = *(longlong **)(lVar7 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 != _Py_NoneStruct_exref) goto LAB_180150f0f;
    pcVar12 = (code *)0x0;
  }
  *(code **)(lVar7 + 0x28) = pcVar12;
  if (((plVar20 != (longlong *)0x0) && (-1 < (int)*plVar20)) &&
     (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
    (**(code **)(plVar20[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_180150f58:
  plVar8 = local_88;
  local_108 = local_98;
  local_110 = local_88;
  local_118 = local_b0;
  local_120 = local_a0;
  local_128 = local_a8;
  local_100 = plVar16;
  local_f8 = plVar9;
  FUN_18036d0a0(plVar13,"ooooooooo",local_b8,local_90);
  plVar20 = DAT_1803b6a88;
  if (plVar13 == DAT_1803b6a88) {
    if ((-1 < (int)*DAT_1803b6a88) && (*DAT_1803b6a88 = *DAT_1803b6a88 + -1, *plVar20 == 0)) {
      (**(code **)(plVar20[1] + 0x30))();
    }
    DAT_1803b6a88 = (longlong *)0x0;
  }
  lVar4 = **(longlong **)(param_1 + 0x38);
  plVar13 = *(longlong **)(lVar4 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar4 + 8);
  *(undefined4 *)(plVar13 + 8) = 0xffffffff;
  plVar20 = (longlong *)plVar13[2];
  if (((plVar20 != (longlong *)0x0) && (plVar13[2] = 0, -1 < (int)*plVar20)) &&
     (*plVar20 = *plVar20 + -1, *plVar20 == 0)) {
    (**(code **)(plVar20[1] + 0x30))();
  }
  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))(plVar13);
  }
  plVar13[10] = 0;
  if (((local_a8 != (longlong *)0x0) && (-1 < (int)*local_a8)) &&
     (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
    (**(code **)(local_a8[1] + 0x30))();
  }
  if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))();
  }
  if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
     (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
    (**(code **)(local_b0[1] + 0x30))();
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*local_b8) && (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
    (**(code **)(local_b8[1] + 0x30))();
  }
  if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))();
  }
  plVar13 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar7;
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))();
  }
  return (int *)0x0;
}
