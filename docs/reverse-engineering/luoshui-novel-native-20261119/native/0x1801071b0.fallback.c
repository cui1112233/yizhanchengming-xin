/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801071b0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

code * FUN_1801071b0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  int *piVar2;
  undefined8 *puVar3;
  undefined8 uVar4;
  char cVar5;
  int iVar6;
  int iVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  code *pcVar12;
  code *pcVar13;
  longlong lVar14;
  longlong *plVar15;
  longlong *plVar16;
  code *pcVar17;
  uint uVar18;
  longlong lVar19;
  undefined8 uVar20;
  undefined1 auStack_238 [32];
  longlong *local_218;
  longlong *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  code *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  code *local_190;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  int local_168;
  code *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  code *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  code *local_90;
  code *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  ulonglong local_48;
  
  plVar11 = DAT_1803b5e90;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_238;
  local_180 = (longlong *)*param_3;
  plVar15 = (longlong *)param_3[1];
  local_148 = (longlong *)param_3[2];
  local_d8 = (longlong *)param_3[3];
  local_a8 = (longlong *)param_3[4];
  local_120 = (longlong *)param_3[5];
  local_c0 = (longlong *)0x0;
  local_b8 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_f0 = (longlong *)0x0;
  local_e8 = (longlong *)0x0;
  local_e0 = (longlong *)0x0;
  local_110 = (longlong *)0x0;
  local_108 = (longlong *)0x0;
  local_100 = (longlong *)0x0;
  local_140 = (longlong *)0x0;
  local_138 = (code *)0x0;
  local_130 = (longlong *)0x0;
  local_128 = (longlong *)0x0;
  local_118 = (longlong *)0x0;
  local_170 = DAT_1803b5e90;
  local_c8 = param_1;
  local_a0 = plVar15;
  if (DAT_1803b5e90 == (longlong *)0x0) {
LAB_180107286:
    DAT_1803b5e90 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3ba0,DAT_1803ca778,0xa0);
  }
  else if ((1 < *DAT_1803b5e90) || (DAT_1803b5e90[2] != 0)) {
    if ((DAT_1803b5e90 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b5e90 && (*DAT_1803b5e90 = *DAT_1803b5e90 + -1, *plVar11 == 0)))) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    goto LAB_180107286;
  }
  plVar11 = DAT_1803b5e90;
  local_170 = DAT_1803b5e90;
  FUN_180001d00(param_1,DAT_1803b5e90);
  uVar4 = DAT_1803ca780;
  uVar20 = DAT_1803b4840;
  pcVar12 = _Py_NoneStruct_exref;
  iVar7 = 0x3ac;
  local_210 = (longlong *)DAT_1803b4258;
  local_218 = (longlong *)DAT_1803b4c60;
  *(undefined4 *)(plVar11 + 5) = 0x3ac;
  plVar8 = (longlong *)FUN_1803717e0(param_1,uVar20,uVar4,pcVar12);
  if (plVar8 == (longlong *)0x0) {
    plVar15 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    if (((code *)plVar8[1] == PyModule_Type_exref) || (iVar6 = PyType_IsSubtype(), iVar6 != 0)) {
      local_218 = (longlong *)DAT_1803ab250;
      plVar16 = (longlong *)FUN_180371d80(param_1,plVar8,DAT_1803ca780,DAT_1803b4850);
    }
    else {
      plVar16 = (longlong *)FUN_180371890(param_1,plVar8,DAT_1803b4850);
    }
    if (plVar16 == (longlong *)0x0) {
      plVar15 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      local_c0 = plVar16;
      if (((code *)plVar8[1] == PyModule_Type_exref) || (iVar6 = PyType_IsSubtype(), iVar6 != 0)) {
        local_218 = (longlong *)DAT_1803ab250;
        plVar9 = (longlong *)FUN_180371d80(param_1,plVar8,DAT_1803ca780,DAT_1803b4c68);
      }
      else {
        plVar9 = (longlong *)FUN_180371890(param_1,plVar8,DAT_1803b4c68);
      }
      if (plVar9 != (longlong *)0x0) {
        local_b8 = plVar9;
        if (((code *)plVar8[1] == PyModule_Type_exref) || (iVar6 = PyType_IsSubtype(), iVar6 != 0))
        {
          local_218 = (longlong *)DAT_1803ab250;
          plVar10 = (longlong *)FUN_180371d80(param_1,plVar8,DAT_1803ca780,DAT_1803b4c70);
        }
        else {
          plVar10 = (longlong *)FUN_180371890(param_1,plVar8,DAT_1803b4c70);
        }
        if (plVar10 != (longlong *)0x0) {
          local_b0 = plVar10;
          if (((code *)plVar8[1] == PyModule_Type_exref) || (iVar6 = PyType_IsSubtype(), iVar6 != 0)
             ) {
            local_218 = (longlong *)DAT_1803ab250;
            local_150 = (longlong *)FUN_180371d80(param_1,plVar8,DAT_1803ca780,DAT_1803b4c78);
          }
          else {
            local_150 = (longlong *)FUN_180371890(param_1,plVar8,DAT_1803b4c78);
          }
          if (local_150 != (longlong *)0x0) {
            local_f0 = local_150;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            *(undefined4 *)(plVar11 + 5) = 0x3af;
            local_f8 = (longlong *)FUN_180376770(param_1,plVar16,plVar15);
            if (local_f8 == (longlong *)0x0) {
              plVar15 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              iVar7 = 0x3af;
            }
            else {
              local_e8 = local_f8;
              plVar8 = (longlong *)FUN_18037f0c0();
              if (plVar8 == (longlong *)0x0) {
LAB_180107538:
                plVar15 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                iVar7 = 0x3b0;
              }
              else {
                iVar7 = FUN_180001510(plVar8);
                if (iVar7 == -1) {
                  plVar15 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  iVar7 = 0x3b0;
                }
                else {
                  if (iVar7 != 1) {
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    plVar8 = DAT_1803b3f30;
                    iVar7 = (int)*DAT_1803b3f30 + 1;
                    if (iVar7 != 0) {
                      *(int *)DAT_1803b3f30 = iVar7;
                    }
                  }
                  plVar15 = (longlong *)FUN_18037f0c0();
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (plVar15 == (longlong *)0x0) goto LAB_180107538;
                  local_168 = 0x3b0;
                  *(undefined4 *)(plVar11 + 5) = 0x3b0;
                  plVar8 = (longlong *)FUN_180376210(param_1,plVar15);
                  if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                  }
                  plVar16 = local_120;
                  plVar15 = DAT_1803b3f30;
                  if (plVar8 == (longlong *)0x0) {
LAB_180107621:
                    plVar15 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    iVar7 = local_168;
                  }
                  else {
                    iVar7 = (int)*DAT_1803b3f30 + 1;
                    iVar6 = (int)*DAT_1803b3f30;
                    if (iVar7 != 0) {
                      *(int *)DAT_1803b3f30 = iVar7;
                      iVar6 = iVar7;
                    }
                    plVar11 = DAT_1803b3f30;
                    local_110 = plVar15;
                    if (iVar6 + 1 != 0) {
                      *(int *)DAT_1803b3f30 = iVar6 + 1;
                    }
                    local_108 = plVar11;
                    local_e0 = plVar8;
                    iVar7 = FUN_18037f2f0(param_1,local_120,DAT_1803b4c80);
                    if (iVar7 == -1) {
                      iVar7 = 0x3b3;
                    }
                    else if (iVar7 == 0) {
LAB_180107770:
                      iVar7 = FUN_18037f2f0(param_1,plVar16,DAT_1803b4c88);
                      if (iVar7 == -1) {
                        iVar7 = 0x3b5;
                      }
                      else if (iVar7 == 0) {
LAB_18010786d:
                        plVar16 = (longlong *)FUN_18037f0c0();
                        if (plVar16 != (longlong *)0x0) {
                          plVar11 = (longlong *)FUN_180024e50(param_1,local_110,DAT_1803b4c98);
                          if (plVar11 == (longlong *)0x0) {
                            plVar15 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                              (**(code **)(plVar16[1] + 0x30))(plVar16);
                            }
                            plVar11 = local_170;
                            iVar7 = 0x3ba;
                          }
                          else {
                            plVar8 = (longlong *)FUN_180024e50(param_1,plVar8,DAT_1803b4c98);
                            if (plVar8 == (longlong *)0x0) {
                              plVar15 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                              {
                                (**(code **)(plVar16[1] + 0x30))(plVar16);
                              }
                              if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0))
                              {
                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                              }
                              plVar11 = local_170;
                              iVar7 = 0x3bb;
                            }
                            else {
                              plVar15 = (longlong *)FUN_180024e50(param_1,local_108,DAT_1803b4c98);
                              if (plVar15 == (longlong *)0x0) {
                                plVar15 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                if ((-1 < (int)*plVar16) &&
                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                  (**(code **)(plVar16[1] + 0x30))(plVar16);
                                }
                                if ((-1 < (int)*plVar11) &&
                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                }
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                }
                                plVar11 = local_170;
                                iVar7 = 0x3bc;
                              }
                              else {
                                local_168 = 0x3b7;
                                *(undefined4 *)(local_170 + 5) = 0x3b7;
                                local_70 = local_148;
                                local_68 = local_d8;
                                local_60 = plVar11;
                                local_58 = plVar8;
                                local_50 = plVar15;
                                plVar9 = (longlong *)
                                         FUN_18037ccb0(param_1,plVar9,&local_70,DAT_1803b4ca0);
                                if ((-1 < (int)*plVar11) &&
                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                }
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                }
                                if ((-1 < (int)*plVar15) &&
                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                }
                                plVar11 = local_170;
                                if (plVar9 == (longlong *)0x0) {
                                  plVar15 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  iVar7 = local_168;
                                  if ((-1 < (int)*plVar16) &&
                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                    (**(code **)(plVar16[1] + 0x30))(plVar16);
                                    plVar11 = local_170;
                                    iVar7 = local_168;
                                  }
                                }
                                else {
                                  *(undefined4 *)(local_170 + 5) = 0x3b7;
                                  plVar15 = (longlong *)FUN_180376770(param_1,plVar16,plVar9);
                                  if ((-1 < (int)*plVar16) &&
                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                    (**(code **)(plVar16[1] + 0x30))(plVar16);
                                  }
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
                                  if (plVar15 == (longlong *)0x0) goto LAB_180107621;
                                  if ((-1 < (int)*plVar15) &&
                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                                  }
                                  plVar8 = (longlong *)
                                           FUN_18036f270(param_1,local_120,DAT_1803b4ca8,
                                                         DAT_1803b3f30);
                                  if (plVar8 == (longlong *)0x0) {
                                    plVar15 = *(longlong **)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    iVar7 = 0x3be;
                                  }
                                  else {
                                    iVar7 = FUN_180001510(plVar8);
                                    if (iVar7 == -1) {
                                      plVar15 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar8) &&
                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                      }
                                      iVar7 = 0x3be;
                                    }
                                    else {
                                      if (iVar7 != 1) {
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                        }
                                        local_100 = DAT_1803b3f30;
                                        iVar7 = (int)*DAT_1803b3f30 + 1;
                                        plVar8 = local_100;
                                        if (iVar7 != 0) {
                                          *(int *)DAT_1803b3f30 = iVar7;
                                        }
                                      }
                                      local_100 = plVar8;
                                      iVar7 = FUN_180001510(local_100);
                                      if (iVar7 == -1) {
                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        iVar7 = 0x3bf;
                                      }
                                      else if (iVar7 == 0) {
LAB_180108713:
                                        plVar11 = (longlong *)FUN_18037f0c0();
                                        piVar1 = DAT_1803b4ce0;
                                        if (plVar11 == (longlong *)0x0) {
                                          plVar15 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          plVar11 = local_170;
                                          iVar7 = 0x3ce;
                                        }
                                        else {
                                          lVar14 = *(longlong *)(param_1 + 0x10);
                                          plVar8 = *(longlong **)(lVar14 + 0x41da8);
                                          if (plVar8 == (longlong *)0x0) {
                                            plVar8 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,3)
                                            ;
                                          }
                                          else {
                                            *(longlong *)(lVar14 + 0x41da8) = plVar8[3];
                                            *(int *)(lVar14 + 0x41e40) =
                                                 *(int *)(lVar14 + 0x41e40) + -1;
                                            *plVar8 = 1;
                                          }
                                          plVar15 = plVar8 + -2;
                                          plVar8[3] = 0;
                                          plVar8[4] = 0;
                                          plVar8[5] = 0;
                                          lVar14 = _PyThreadState_GetCurrent();
                                          plVar16 = local_148;
                                          lVar14 = *(longlong *)
                                                    (*(longlong *)(lVar14 + 0x10) + 0xd0);
                                          puVar3 = *(undefined8 **)(lVar14 + 8);
                                          *puVar3 = plVar15;
                                          plVar8[-1] = plVar8[-1] & 3;
                                          plVar8[-1] = plVar8[-1] | (ulonglong)puVar3;
                                          *plVar15 = lVar14;
                                          *(longlong **)(lVar14 + 8) = plVar15;
                                          if (*piVar1 + 1 != 0) {
                                            *piVar1 = *piVar1 + 1;
                                          }
                                          plVar8[3] = (longlong)piVar1;
                                          if (local_148 == (longlong *)0x0) {
                                            plVar16 = (longlong *)
                                                      PyUnicode_FromFormat
                                                                (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b4698);
                                            plVar15 = (longlong *)
                                                      FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                          
                                                  PyExc_UnboundLocalError_exref,plVar16);
                                            if ((-1 < (int)*plVar16) &&
                                               (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                              (**(code **)(plVar16[1] + 0x30))(plVar16);
                                            }
                                            FUN_18000d1d0(param_1,plVar15);
                                          }
                                          else {
                                            lVar14 = FUN_180375520(param_1,local_148,DAT_1803b3f30);
                                            if (lVar14 != 0) {
                                              plVar8[4] = lVar14;
                                              piVar1 = DAT_1803b4aa8;
                                              if (*DAT_1803b4aa8 + 1 != 0) {
                                                *DAT_1803b4aa8 = *DAT_1803b4aa8 + 1;
                                              }
                                              plVar8[5] = (longlong)piVar1;
                                              plVar15 = (longlong *)
                                                        PyUnicode_Join(DAT_1803b3f30,plVar8);
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                (**(code **)(plVar8[1] + 0x30))(plVar8);
                                              }
                                              if (plVar15 != (longlong *)0x0) {
                                                iVar7 = 0x3ce;
                                                *(undefined4 *)(local_170 + 5) = 0x3ce;
                                                plVar8 = (longlong *)
                                                         FUN_180376770(param_1,plVar11,plVar15);
                                                if ((-1 < (int)*plVar11) &&
                                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                }
                                                if ((-1 < (int)*plVar15) &&
                                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                }
                                                if (plVar8 != (longlong *)0x0) {
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                  }
                                                  goto LAB_180108d6b;
                                                }
                                                goto LAB_1801089bf;
                                              }
                                              plVar15 = *(longlong **)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              if ((-1 < (int)*plVar11) &&
                                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                                              }
                                              plVar11 = local_170;
                                              iVar7 = 0x3ce;
                                              goto LAB_1801089d3;
                                            }
                                            plVar15 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                          }
                                          iVar7 = 0x3ce;
                                          if ((-1 < (int)*plVar11) &&
                                             (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                            (**(code **)(plVar11[1] + 0x30))(plVar11);
                                          }
                                          plVar11 = local_170;
                                          if ((-1 < (int)*plVar8) &&
                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                                            plVar11 = local_170;
                                          }
                                        }
                                      }
                                      else {
                                        plVar15 = (longlong *)FUN_18037f0c0();
                                        iVar7 = 0x3c0;
                                        if (plVar15 == (longlong *)0x0) {
                                          plVar15 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                        }
                                        else {
                                          *(undefined4 *)(plVar11 + 5) = 0x3c0;
                                          local_168 = 0x3c0;
                                          plVar8 = (longlong *)
                                                   FUN_1803770d0(param_1,plVar15,DAT_1803b4cb0);
                                          if ((-1 < (int)*plVar15) &&
                                             (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                            (**(code **)(plVar15[1] + 0x30))(plVar15);
                                          }
                                          if (plVar8 != (longlong *)0x0) {
                                            plVar11 = (longlong *)FUN_180001930();
                                            local_d0 = plVar11;
                                            if ((-1 < (int)*plVar8) &&
                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                                            }
                                            if (plVar11 != (longlong *)0x0) {
LAB_180107c60:
                                              do {
                                                plVar9 = local_118;
                                                plVar16 = local_130;
                                                pcVar13 = local_138;
                                                plVar8 = local_140;
                                                pcVar12 = *(code **)(plVar11[1] + 0xe0);
                                                if (pcVar12 == (code *)0x0) {
                                                  PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                                               "\'%s\' object is not an iterator",
                                                               *(undefined8 *)(plVar11[1] + 0x18));
LAB_18010869b:
                                                  plVar15 = *(longlong **)(param_1 + 0x60);
                                                  if (plVar15 != (longlong *)0x0) {
                                                    cVar5 = FUN_1800014b0();
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    if (cVar5 == '\0') {
                                                      iVar7 = local_168;
                                                      local_140 = plVar8;
                                                      local_138 = pcVar13;
                                                      local_130 = plVar16;
                                                      local_118 = plVar9;
                                                      if (plVar9 != (longlong *)0x0)
                                                      goto LAB_180108766;
                                                      goto LAB_180108782;
                                                    }
                                                    if ((-1 < (int)*plVar15) &&
                                                       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                      (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                    }
                                                  }
                                                  if (((plVar9 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar9)) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                  }
                                                  goto LAB_180108713;
                                                }
                                                plVar15 = (longlong *)(*pcVar12)(plVar11);
                                                plVar11 = local_d0;
                                                if (plVar15 == (longlong *)0x0) goto LAB_18010869b;
                                                local_118 = plVar15;
                                                if (((plVar9 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar9)) &&
                                                   (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                }
                                                iVar7 = (int)*plVar15 + 1;
                                                if (iVar7 != 0) {
                                                  *(int *)plVar15 = iVar7;
                                                }
                                                local_140 = plVar15;
                                                if (((plVar8 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar8)) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  (**(code **)(plVar8[1] + 0x30))();
                                                }
                                                plVar11 = (longlong *)FUN_18037f0c0();
                                                plVar8 = local_170;
                                                if (plVar11 == (longlong *)0x0) {
LAB_180108663:
                                                  plVar15 = *(longlong **)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  iVar7 = 0x3c1;
                                                  goto LAB_180108766;
                                                }
                                                *(undefined4 *)(local_170 + 5) = 0x3c1;
                                                plVar9 = (longlong *)FUN_180376210(param_1,plVar11);
                                                if ((-1 < (int)*plVar11) &&
                                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                }
                                                if (plVar9 == (longlong *)0x0) goto LAB_180108663;
                                                plVar11 = (longlong *)FUN_18037f0c0();
                                                if ((-1 < (int)*plVar9) &&
                                                   (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                }
                                                if (plVar11 == (longlong *)0x0) goto LAB_180108663;
                                                *(undefined4 *)(plVar8 + 5) = 0x3c1;
                                                plVar9 = (longlong *)
                                                         FUN_1803770d0(param_1,plVar11,DAT_1803b4cb8
                                                                      );
                                                local_178 = plVar9;
                                                if ((-1 < (int)*plVar11) &&
                                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                }
                                                if (plVar9 == (longlong *)0x0) goto LAB_180108663;
                                                local_140 = plVar9;
                                                if ((-1 < (int)*plVar15) &&
                                                   (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                }
                                                iVar7 = FUN_180001510(plVar9);
                                                if (iVar7 == -1) {
                                                  plVar15 = *(longlong **)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  iVar7 = 0x3c2;
                                                  goto LAB_180108766;
                                                }
                                                plVar11 = local_d0;
                                              } while (iVar7 == 0);
                                              plVar15 = (longlong *)
                                                        FUN_180024e50(param_1,plVar9,DAT_1803b4cc0);
                                              if (plVar15 == (longlong *)0x0) {
LAB_180108635:
                                                plVar15 = *(longlong **)(param_1 + 0x60);
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                iVar7 = 0x3c4;
                                                goto LAB_180108766;
                                              }
                                              plVar11 = (longlong *)FUN_18037f0c0();
                                              if ((-1 < (int)*plVar15) &&
                                                 (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                (**(code **)(plVar15[1] + 0x30))(plVar15);
                                              }
                                              if (plVar11 == (longlong *)0x0) goto LAB_180108635;
                                              *(undefined4 *)(plVar8 + 5) = 0x3c4;
                                              local_160 = (code *)FUN_180376210(param_1,plVar11);
                                              if ((-1 < (int)*plVar11) &&
                                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                                              }
                                              if (local_160 == (code *)0x0) goto LAB_180108635;
                                              local_138 = local_160;
                                              if (((pcVar13 != (code *)0x0) &&
                                                  (-1 < *(int *)pcVar13)) &&
                                                 (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                                                 *(longlong *)pcVar13 == 0)) {
                                                (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))
                                                          (pcVar13);
                                              }
                                              if (local_160 != _Py_TrueStruct_exref) {
                                                if ((local_160 != _Py_FalseStruct_exref) &&
                                                   (local_160 != _Py_NoneStruct_exref)) {
                                                  lVar14 = *(longlong *)(local_160 + 8);
                                                  if ((*(longlong *)(lVar14 + 0x60) == 0) ||
                                                     (pcVar12 = *(code **)(*(longlong *)
                                                                            (lVar14 + 0x60) + 0x48),
                                                     pcVar12 == (code *)0x0)) {
                                                    if (((*(longlong **)(lVar14 + 0x70) ==
                                                          (longlong *)0x0) ||
                                                        (pcVar12 = (code *)**(longlong **)
                                                                             (lVar14 + 0x70),
                                                        pcVar12 == (code *)0x0)) &&
                                                       ((*(longlong **)(lVar14 + 0x68) ==
                                                         (longlong *)0x0 ||
                                                        (pcVar12 = (code *)**(longlong **)
                                                                             (lVar14 + 0x68),
                                                        pcVar12 == (code *)0x0))))
                                                    goto LAB_180107f2d;
                                                    lVar14 = (*pcVar12)(local_160);
                                                  }
                                                  else {
                                                    iVar7 = (*pcVar12)(local_160);
                                                    lVar14 = (longlong)iVar7;
                                                  }
                                                  if (0 < lVar14) goto LAB_180107f2d;
                                                  if (lVar14 != 0) {
                                                    plVar15 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    iVar7 = 0x3c5;
                                                    goto LAB_180108766;
                                                  }
                                                }
LAB_1801083a1:
                                                lVar14 = *(longlong *)(param_1 + 0x10);
                                                if (((*(int *)(*(longlong *)(lVar14 + 0x60) + 0x4d4)
                                                      != 0) || (*(int *)(lVar14 + 0x1a0) != 0)) &&
                                                   ((iVar7 = Py_MakePendingCalls(), iVar7 < 0 &&
                                                    (*(longlong *)(param_1 + 0x60) != 0)))) {
                                                  plVar15 = *(longlong **)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  iVar7 = local_168;
                                                  goto LAB_180108766;
                                                }
                                                if (*(int *)(lVar14 + 0x174) != 0) {
                                                  PyEval_SaveThread();
                                                  PyEval_AcquireThread(param_1);
                                                }
                                                plVar11 = local_d0;
                                                if (*(longlong *)(param_1 + 0x80) != 0)
                                                goto LAB_180108511;
                                                goto LAB_180107c60;
                                              }
LAB_180107f2d:
                                              lVar14 = local_f8[1];
                                              pcVar12 = *(code **)(lVar14 + 0x90);
                                              lVar19 = DAT_1803b4858;
                                              if (pcVar12 == (code *)0x0) {
                                                pcVar12 = *(code **)(lVar14 + 0x40);
                                                uVar18 = *(uint *)(DAT_1803b4858 + 0x20) & 0x60;
                                                if (pcVar12 != (code *)0x0) {
                                                  lVar19 = DAT_1803b4858 + 0x28;
                                                  if (uVar18 != 0x60) {
                                                    lVar19 = *(longlong *)(DAT_1803b4858 + 0x30);
                                                  }
                                                  goto LAB_180107f6d;
                                                }
                                                lVar19 = DAT_1803b4858 + 0x28;
                                                if (uVar18 != 0x60) {
                                                  lVar19 = *(longlong *)(DAT_1803b4858 + 0x30);
                                                }
                                                PyErr_Format(*(undefined8 *)
                                                              PyExc_AttributeError_exref,
                                                             "\'%s\' object has no attribute \'%s\'"
                                                             ,*(undefined8 *)(lVar14 + 0x18),lVar19)
                                                ;
                                              }
                                              else {
LAB_180107f6d:
                                                plVar15 = (longlong *)(*pcVar12)(local_f8,lVar19);
                                                if (plVar15 != (longlong *)0x0) {
                                                  if (local_148 == (longlong *)0x0) {
                                                    if ((-1 < (int)*plVar15) &&
                                                       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                      (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                    }
                                                    plVar11 = (longlong *)
                                                              PyUnicode_FromFormat
                                                                        (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b4698);
                                                  plVar15 = (longlong *)
                                                            FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                                      
                                                  PyExc_UnboundLocalError_exref,plVar11);
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                  }
                                                  FUN_18000d1d0(param_1,plVar15);
                                                  iVar7 = 0x3c6;
                                                  goto LAB_180108766;
                                                  }
                                                  *(undefined4 *)(plVar8 + 5) = 0x3c6;
                                                  plVar11 = (longlong *)
                                                            FUN_180376770(param_1,plVar15,local_148)
                                                  ;
                                                  if ((-1 < (int)*plVar15) &&
                                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                  }
                                                  if (plVar11 != (longlong *)0x0) {
                                                    local_130 = plVar11;
                                                    if (((plVar16 != (longlong *)0x0) &&
                                                        (-1 < (int)*plVar16)) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                    }
                                                    plVar11 = (longlong *)FUN_180001930();
                                                    if (plVar11 == (longlong *)0x0) {
                                                      plVar15 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_180108576:
                                                      iVar7 = 0x3c7;
                                                      goto LAB_180108766;
                                                    }
                                                    local_158 = (longlong *)PySet_New(0);
                                                    plVar8 = (longlong *)0x0;
                                                    while( true ) {
                                                      pcVar12 = *(code **)(plVar11[1] + 0xe0);
                                                      if (pcVar12 == (code *)0x0) break;
                                                      plVar15 = (longlong *)(*pcVar12)(plVar11);
                                                      if (plVar15 == (longlong *)0x0)
                                                      goto LAB_18010818f;
                                                      if (((plVar8 != (longlong *)0x0) &&
                                                          (-1 < (int)*plVar8)) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                      }
                                                      iVar7 = (int)*plVar15 + 1;
                                                      if (iVar7 != 0) {
                                                        *(int *)plVar15 = iVar7;
                                                      }
                                                      if (((plVar8 != (longlong *)0x0) &&
                                                          (-1 < (int)*plVar8)) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        (**(code **)(plVar8[1] + 0x30))();
                                                      }
                                                      plVar16 = (longlong *)FUN_18037f0c0();
                                                      plVar9 = local_158;
                                                      plVar8 = plVar15;
                                                      if (plVar16 == (longlong *)0x0) {
LAB_180108410:
                                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                                        goto LAB_18010841a;
                                                      }
                                                      iVar7 = PySet_Add(local_158,plVar16);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                      }
                                                      if (iVar7 == -1) goto LAB_180108410;
                                                      lVar14 = *(longlong *)(param_1 + 0x10);
                                                      if ((((*(int *)(*(longlong *)(lVar14 + 0x60) +
                                                                     0x4d4) != 0) ||
                                                           (*(int *)(lVar14 + 0x1a0) != 0)) &&
                                                          (iVar7 = Py_MakePendingCalls(), iVar7 < 0)
                                                          ) && (*(longlong *)(param_1 + 0x60) != 0))
                                                      {
                                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        goto LAB_180108422;
                                                      }
                                                      if (*(int *)(lVar14 + 0x174) != 0) {
                                                        PyEval_SaveThread();
                                                        PyEval_AcquireThread(param_1);
                                                      }
                                                      if (*(longlong *)(param_1 + 0x80) != 0) {
                                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x80) = 0;
                                                        uVar20 = FUN_180373490(param_1);
                                                        *(undefined8 *)(param_1 + 0x60) = uVar20;
                                                        if (((plVar15 != (longlong *)0x0) &&
                                                            (-1 < (int)*plVar15)) &&
                                                           (*plVar15 = *plVar15 + -1, *plVar15 == 0)
                                                           ) {
                                                          (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                        }
                                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        goto LAB_180108422;
                                                      }
                                                    }
                                                    PyErr_Format(*(undefined8 *)
                                                                  PyExc_TypeError_exref,
                                                                 "\'%s\' object is not an iterator",
                                                                 *(undefined8 *)(plVar11[1] + 0x18))
                                                    ;
LAB_18010818f:
                                                    plVar15 = *(longlong **)(param_1 + 0x60);
                                                    if (plVar15 != (longlong *)0x0) {
                                                      plVar16 = plVar15;
                                                      if ((*(uint *)((longlong *)plVar15[1] + 0x15)
                                                          & 0x40000000) != 0) {
                                                        plVar16 = (longlong *)plVar15[1];
                                                      }
                                                      if ((plVar16 !=
                                                           *(longlong **)PyExc_StopIteration_exref)
                                                         && (((plVar9 = local_158,
                                                              -1 < *(int *)(plVar16[1] + 0xa8) ||
                                                              ((*(uint *)(plVar16 + 0x15) &
                                                               0x40000000) == 0)) ||
                                                             (cVar5 = FUN_180382640(),
                                                             plVar9 = local_158, cVar5 != '\x01'))))
                                                      {
LAB_18010841a:
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
LAB_180108422:
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                        }
                                                        if (((plVar8 != (longlong *)0x0) &&
                                                            (-1 < (int)*plVar8)) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                        }
                                                        if (((plVar8 == (longlong *)0x0) ||
                                                            ((int)*plVar8 < 0)) ||
                                                           (*plVar8 = *plVar8 + -1, *plVar8 != 0))
                                                        goto LAB_180108576;
                                                        (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                        iVar7 = 0x3c7;
                                                        goto LAB_180108766;
                                                      }
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                      if ((-1 < (int)*plVar15) &&
                                                         (*plVar15 = *plVar15 + -1, *plVar15 == 0))
                                                      {
                                                        (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                      }
                                                    }
                                                    plVar15 = local_158;
                                                    iVar7 = (int)*local_158 + 1;
                                                    if (iVar7 != 0) {
                                                      *(int *)local_158 = iVar7;
                                                    }
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      (**(code **)(plVar11[1] + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar15) &&
                                                       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                      (**(code **)(plVar15[1] + 0x30))(plVar15);
                                                    }
                                                    if (((plVar8 != (longlong *)0x0) &&
                                                        (-1 < (int)*plVar8)) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                    }
                                                    if (((plVar8 != (longlong *)0x0) &&
                                                        (-1 < (int)*plVar8)) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                    }
                                                    plVar11 = local_128;
                                                    local_128 = plVar15;
                                                    if (((plVar11 != (longlong *)0x0) &&
                                                        (-1 < (int)*plVar11)) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      (**(code **)(plVar11[1] + 0x30))();
                                                    }
                                                    pcVar12 = local_160;
                                                    iVar7 = PySequence_Contains(plVar15,local_160);
                                                    if (iVar7 == -1) {
                                                      plVar15 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                      iVar7 = 0x3c8;
                                                      goto LAB_180108766;
                                                    }
                                                    if (iVar7 == 0) {
                                                      plVar11 = (longlong *)FUN_18037f0c0();
                                                      plVar15 = local_170;
                                                      if (plVar11 == (longlong *)0x0) {
                                                        plVar15 = *(longlong **)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        iVar7 = 0x3c9;
                                                      }
                                                      else {
                                                        local_80 = local_148;
                                                        *(undefined4 *)(local_170 + 5) = 0x3c9;
                                                        local_90 = pcVar12;
                                                        local_88 = pcVar12;
                                                        local_78 = local_178;
                                                        plVar8 = (longlong *)
                                                                 FUN_18037ccb0(param_1,local_150,
                                                                               &local_90,
                                                                               DAT_1803b4cd8);
                                                        if (plVar8 == (longlong *)0x0) {
                                                          plVar15 = *(longlong **)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          if ((-1 < (int)*plVar11) &&
                                                             (*plVar11 = *plVar11 + -1,
                                                             *plVar11 == 0)) {
                                                            (**(code **)(plVar11[1] + 0x30))
                                                                      (plVar11);
                                                          }
                                                          iVar7 = 0x3c9;
                                                        }
                                                        else {
                                                          *(undefined4 *)(plVar15 + 5) = 0x3c9;
                                                          plVar15 = (longlong *)
                                                                    FUN_180376770(param_1,plVar11,
                                                                                  plVar8);
                                                          if ((-1 < (int)*plVar11) &&
                                                             (*plVar11 = *plVar11 + -1,
                                                             *plVar11 == 0)) {
                                                            (**(code **)(plVar11[1] + 0x30))
                                                                      (plVar11);
                                                          }
                                                          if ((-1 < (int)*plVar8) &&
                                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                                          {
                                                            (**(code **)(plVar8[1] + 0x30))(plVar8);
                                                          }
                                                          if (plVar15 != (longlong *)0x0) {
                                                            if ((-1 < (int)*plVar15) &&
                                                               (*plVar15 = *plVar15 + -1,
                                                               *plVar15 == 0)) {
                                                              (**(code **)(plVar15[1] + 0x30))
                                                                        (plVar15);
                                                            }
                                                            goto LAB_1801083a1;
                                                          }
                                                          plVar15 = *(longlong **)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          iVar7 = 0x3c9;
                                                        }
                                                      }
                                                      goto LAB_180108766;
                                                    }
                                                    goto LAB_1801083a1;
                                                  }
                                                }
                                              }
                                              plVar15 = *(longlong **)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              iVar7 = 0x3c6;
                                              goto LAB_180108766;
                                            }
                                            goto LAB_1801089bf;
                                          }
                                          plVar15 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                          goto LAB_1801089d3;
                        }
                        iVar7 = 0x3b7;
                      }
                      else {
                        plVar16 = (longlong *)FUN_18037f0c0();
                        if (plVar16 != (longlong *)0x0) {
                          iVar7 = FUN_180001510(plVar16);
                          if (iVar7 == -1) {
                            plVar15 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                              (**(code **)(plVar16[1] + 0x30))(plVar16);
                            }
                            plVar11 = local_170;
                            iVar7 = 0x3b6;
                            goto LAB_1801089d3;
                          }
                          if (iVar7 != 1) {
                            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                              (**(code **)(plVar16[1] + 0x30))(plVar16);
                            }
                            plVar16 = DAT_1803b3f30;
                            iVar7 = (int)*DAT_1803b3f30 + 1;
                            if (iVar7 != 0) {
                              *(int *)DAT_1803b3f30 = iVar7;
                            }
                          }
                          plVar15 = (longlong *)FUN_1803749a0(plVar16);
                          if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                            (**(code **)(plVar16[1] + 0x30))(plVar16);
                          }
                          if (plVar15 != (longlong *)0x0) {
                            local_108 = plVar15;
                            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                              (**(code **)(plVar11[1] + 0x30))(plVar11);
                            }
                            goto LAB_18010786d;
                          }
                        }
                        iVar7 = 0x3b6;
                      }
                    }
                    else {
                      plVar16 = (longlong *)FUN_18037f0c0();
                      if (plVar16 != (longlong *)0x0) {
                        iVar7 = FUN_180001510(plVar16);
                        if (iVar7 == -1) {
                          plVar15 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                            (**(code **)(plVar16[1] + 0x30))(plVar16);
                          }
                          plVar11 = local_170;
                          iVar7 = 0x3b4;
                          goto LAB_1801089d3;
                        }
                        if (iVar7 != 1) {
                          if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                            (**(code **)(plVar16[1] + 0x30))(plVar16);
                          }
                          plVar16 = DAT_1803b3f30;
                          iVar7 = (int)*DAT_1803b3f30 + 1;
                          if (iVar7 != 0) {
                            *(int *)DAT_1803b3f30 = iVar7;
                          }
                        }
                        plVar10 = (longlong *)FUN_1803749a0(plVar16);
                        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                        }
                        if (plVar10 != (longlong *)0x0) {
                          plVar16 = local_120;
                          local_110 = plVar10;
                          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                            (**(code **)(plVar15[1] + 0x30))(plVar15);
                            plVar16 = local_120;
                          }
                          goto LAB_180107770;
                        }
                      }
                      iVar7 = 0x3b4;
                    }
LAB_1801089bf:
                    plVar15 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    plVar11 = local_170;
                  }
                }
              }
            }
            goto LAB_1801089d3;
          }
        }
      }
      plVar15 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
  }
  goto LAB_1801089d3;
LAB_180108511:
  plVar15 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x80) = 0;
  uVar20 = FUN_180373490(param_1);
  *(undefined8 *)(param_1 + 0x60) = uVar20;
  if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
    (**(code **)(plVar15[1] + 0x30))(plVar15);
  }
  plVar15 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x60) = 0;
  iVar7 = local_168;
LAB_180108766:
  if ((-1 < (int)*local_118) && (*local_118 = *local_118 + -1, *local_118 == 0)) {
    (**(code **)(local_118[1] + 0x30))(local_118);
  }
LAB_180108782:
  plVar11 = local_170;
  if ((-1 < (int)*local_d0) && (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
    plVar11 = local_170;
  }
LAB_1801089d3:
  piVar1 = *(int **)(param_1 + 0xf8);
  local_150 = (longlong *)0x0;
  if ((piVar1 != (int *)0x0) && (*piVar1 + 1 != 0)) {
    *piVar1 = *piVar1 + 1;
  }
  piVar2 = (int *)plVar15[5];
  if (piVar2 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar11,iVar7);
    plVar11 = (longlong *)plVar15[5];
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_180108a4b:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
LAB_180108a59:
    plVar15[5] = (longlong)pcVar13;
    if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
       (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
      (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
    }
  }
  else if (iVar7 != 0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar11,iVar7);
    *(int **)(pcVar12 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar11 = (longlong *)plVar15[5];
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_180108a4b;
    pcVar13 = (code *)0x0;
    goto LAB_180108a59;
  }
  plVar11 = *(longlong **)(param_1 + 0xf8);
  *(longlong **)(param_1 + 0xf8) = plVar15;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  iVar7 = FUN_18000ce90(param_1,*(undefined8 *)(*(longlong *)(param_1 + 0xf8) + 8),
                        *(undefined8 *)PyExc_Exception_exref);
  if (iVar7 == 0) {
    cVar5 = FUN_180373290(param_1,&local_150);
    uVar20 = 0x3ab;
    if (cVar5 != '\0') {
      uVar20 = 0;
    }
    lVar14 = local_150[5];
    plVar15 = local_150;
    if ((lVar14 != 0) && (*(longlong **)(lVar14 + 0x18) == local_170)) {
      *(undefined4 *)(local_170 + 5) = *(undefined4 *)(lVar14 + 0x24);
    }
  }
  else {
    plVar11 = *(longlong **)(param_1 + 0xf8);
    iVar7 = (int)*plVar11 + 1;
    if (iVar7 != 0) {
      *(int *)plVar11 = iVar7;
    }
    plVar8 = (longlong *)FUN_18037f0c0();
    piVar2 = DAT_1803b4ce8;
    if (plVar8 == (longlong *)0x0) {
      plVar15 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar14 = *(longlong *)(param_1 + 0x10);
      plVar16 = *(longlong **)(lVar14 + 0x41da0);
      if (plVar16 == (longlong *)0x0) {
        plVar16 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
      }
      else {
        *(longlong *)(lVar14 + 0x41da0) = plVar16[3];
        *(int *)(lVar14 + 0x41e3c) = *(int *)(lVar14 + 0x41e3c) + -1;
        *plVar16 = 1;
      }
      plVar15 = plVar16 + -2;
      plVar16[3] = 0;
      plVar16[4] = 0;
      lVar14 = _PyThreadState_GetCurrent();
      lVar14 = *(longlong *)(*(longlong *)(lVar14 + 0x10) + 0xd0);
      puVar3 = *(undefined8 **)(lVar14 + 8);
      *puVar3 = plVar15;
      plVar16[-1] = plVar16[-1] & 3;
      plVar16[-1] = plVar16[-1] | (ulonglong)puVar3;
      *plVar15 = lVar14;
      *(longlong **)(lVar14 + 8) = plVar15;
      if (*piVar2 + 1 != 0) {
        *piVar2 = *piVar2 + 1;
      }
      plVar16[3] = (longlong)piVar2;
      lVar14 = FUN_180375520(param_1,plVar11,DAT_1803b3f30);
      if (lVar14 == 0) {
        plVar15 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
          (**(code **)(plVar16[1] + 0x30))(plVar16);
        }
      }
      else {
        plVar16[4] = lVar14;
        plVar15 = (longlong *)PyUnicode_Join(DAT_1803b3f30,plVar16);
        if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
          (**(code **)(plVar16[1] + 0x30))(plVar16);
        }
        param_1 = local_c8;
        if (plVar15 == (longlong *)0x0) {
          plVar15 = *(longlong **)(local_c8 + 0x60);
          *(undefined8 *)(local_c8 + 0x60) = 0;
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
        }
        else {
          *(undefined4 *)(local_170 + 5) = 0x3d0;
          plVar16 = (longlong *)FUN_180376770(local_c8,plVar8,plVar15);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
            (**(code **)(plVar15[1] + 0x30))(plVar15);
          }
          if (plVar16 != (longlong *)0x0) {
            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
               (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            param_1 = local_c8;
            plVar15 = *(longlong **)(local_c8 + 0xf8);
            *(int **)(local_c8 + 0xf8) = piVar1;
            plVar16 = local_148;
            if (((plVar15 != (longlong *)0x0) && (-1 < (int)*plVar15)) &&
               (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
              (**(code **)(plVar15[1] + 0x30))();
              plVar16 = local_148;
            }
LAB_180108d6b:
            FUN_180001d80(param_1);
            pcVar12 = _Py_NoneStruct_exref;
            if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
              *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
            }
            if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
               (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
              (**(code **)(local_c0[1] + 0x30))();
            }
            if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
               (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
              (**(code **)(local_b8[1] + 0x30))();
            }
            if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
               (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
              (**(code **)(local_b0[1] + 0x30))();
            }
            if (((local_f0 != (longlong *)0x0) && (-1 < (int)*local_f0)) &&
               (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
              (**(code **)(local_f0[1] + 0x30))();
            }
            if (((local_e8 != (longlong *)0x0) && (-1 < (int)*local_e8)) &&
               (*local_e8 = *local_e8 + -1, *local_e8 == 0)) {
              (**(code **)(local_e8[1] + 0x30))();
            }
            if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
               (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
              (**(code **)(local_e0[1] + 0x30))();
            }
            if (((local_110 != (longlong *)0x0) && (-1 < (int)*local_110)) &&
               (*local_110 = *local_110 + -1, *local_110 == 0)) {
              (**(code **)(local_110[1] + 0x30))();
            }
            if (((local_108 != (longlong *)0x0) && (-1 < (int)*local_108)) &&
               (*local_108 = *local_108 + -1, *local_108 == 0)) {
              (**(code **)(local_108[1] + 0x30))();
            }
            if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
               (*local_100 = *local_100 + -1, *local_100 == 0)) {
              (**(code **)(local_100[1] + 0x30))();
            }
            if (((local_140 != (longlong *)0x0) && (-1 < (int)*local_140)) &&
               (*local_140 = *local_140 + -1, *local_140 == 0)) {
              (**(code **)(local_140[1] + 0x30))();
            }
            if (((local_138 != (code *)0x0) && (-1 < *(int *)local_138)) &&
               (*(longlong *)local_138 = *(longlong *)local_138 + -1, *(longlong *)local_138 == 0))
            {
              (**(code **)(*(longlong *)(local_138 + 8) + 0x30))();
            }
            if (((local_130 != (longlong *)0x0) && (-1 < (int)*local_130)) &&
               (*local_130 = *local_130 + -1, *local_130 == 0)) {
              (**(code **)(local_130[1] + 0x30))();
            }
            if (((local_128 != (longlong *)0x0) && (-1 < (int)*local_128)) &&
               (*local_128 = *local_128 + -1, *local_128 == 0)) {
              (**(code **)(local_128[1] + 0x30))();
            }
            if ((-1 < (int)*local_180) && (*local_180 = *local_180 + -1, *local_180 == 0)) {
              (**(code **)(local_180[1] + 0x30))();
            }
            if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
              (**(code **)(local_a0[1] + 0x30))();
            }
            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if ((-1 < (int)*local_d8) && (*local_d8 = *local_d8 + -1, *local_d8 == 0)) {
              (**(code **)(local_d8[1] + 0x30))();
            }
            if ((-1 < (int)*local_a8) && (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
              (**(code **)(local_a8[1] + 0x30))();
            }
            if ((int)*local_120 < 0) {
              return pcVar12;
            }
            *local_120 = *local_120 + -1;
            if (*local_120 != 0) {
              return pcVar12;
            }
            (**(code **)(local_120[1] + 0x30))(local_120);
            return pcVar12;
          }
          plVar15 = *(longlong **)(local_c8 + 0x60);
          *(undefined8 *)(local_c8 + 0x60) = 0;
          param_1 = local_c8;
        }
      }
    }
    uVar20 = 0x3d0;
    if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
  }
  pcVar12 = (code *)0x0;
  plVar11 = *(longlong **)(param_1 + 0xf8);
  *(int **)(param_1 + 0xf8) = piVar1;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  piVar1 = (int *)plVar15[5];
  pcVar13 = pcVar12;
  if (piVar1 == (int *)0x0) {
    pcVar17 = (code *)FUN_18038e2a0(local_170,uVar20);
    plVar11 = (longlong *)plVar15[5];
    if (pcVar17 != _Py_NoneStruct_exref) {
LAB_180109097:
      pcVar13 = pcVar17;
      if ((pcVar17 != (code *)0x0) && (*(int *)pcVar17 + 1 != 0)) {
        *(int *)pcVar17 = *(int *)pcVar17 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == local_170) goto LAB_1801090e0;
    pcVar17 = (code *)FUN_18038e2a0(local_170,uVar20);
    *(int **)(pcVar17 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar11 = (longlong *)plVar15[5];
    if (pcVar17 != _Py_NoneStruct_exref) goto LAB_180109097;
  }
  plVar15[5] = (longlong)pcVar13;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
LAB_1801090e0:
  plVar9 = local_128;
  plVar16 = local_130;
  pcVar13 = local_138;
  plVar8 = local_140;
  local_198 = local_128;
  local_1a0 = local_130;
  local_1a8 = local_138;
  local_1b0 = local_140;
  local_1b8 = local_100;
  local_1c0 = local_108;
  local_1c8 = local_110;
  local_1d0 = local_e0;
  local_1d8 = local_e8;
  local_1e0 = local_f0;
  local_1e8 = local_b0;
  local_1f0 = local_b8;
  local_1f8 = local_c0;
  local_200 = local_120;
  local_208 = local_a8;
  local_210 = local_d8;
  local_218 = local_148;
  local_190 = pcVar12;
  FUN_18036d0a0(local_170,"oooooooooooooooooooo",local_180,local_a0);
  plVar11 = DAT_1803b5e90;
  if (local_170 == DAT_1803b5e90) {
    if ((-1 < (int)*DAT_1803b5e90) && (*DAT_1803b5e90 = *DAT_1803b5e90 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    DAT_1803b5e90 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
     (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
    (**(code **)(local_c0[1] + 0x30))();
  }
  if (((local_b8 != (longlong *)0x0) && (-1 < (int)*local_b8)) &&
     (*local_b8 = *local_b8 + -1, *local_b8 == 0)) {
    (**(code **)(local_b8[1] + 0x30))();
  }
  if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
     (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
    (**(code **)(local_b0[1] + 0x30))();
  }
  if (((local_f0 != (longlong *)0x0) && (-1 < (int)*local_f0)) &&
     (*local_f0 = *local_f0 + -1, *local_f0 == 0)) {
    (**(code **)(local_f0[1] + 0x30))();
  }
  if (((local_e8 != (longlong *)0x0) && (-1 < (int)*local_e8)) &&
     (*local_e8 = *local_e8 + -1, *local_e8 == 0)) {
    (**(code **)(local_e8[1] + 0x30))();
  }
  if (((local_e0 != (longlong *)0x0) && (-1 < (int)*local_e0)) &&
     (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
    (**(code **)(local_e0[1] + 0x30))();
  }
  if (((local_110 != (longlong *)0x0) && (-1 < (int)*local_110)) &&
     (*local_110 = *local_110 + -1, *local_110 == 0)) {
    (**(code **)(local_110[1] + 0x30))();
  }
  if (((local_108 != (longlong *)0x0) && (-1 < (int)*local_108)) &&
     (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
     (*local_100 = *local_100 + -1, *local_100 == 0)) {
    (**(code **)(local_100[1] + 0x30))();
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if ((-1 < (int)*local_180) && (*local_180 = *local_180 + -1, *local_180 == 0)) {
    (**(code **)(local_180[1] + 0x30))();
  }
  if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))();
  }
  if ((-1 < (int)*local_148) && (*local_148 = *local_148 + -1, *local_148 == 0)) {
    (**(code **)(local_148[1] + 0x30))();
  }
  if ((-1 < (int)*local_d8) && (*local_d8 = *local_d8 + -1, *local_d8 == 0)) {
    (**(code **)(local_d8[1] + 0x30))();
  }
  if ((-1 < (int)*local_a8) && (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
    (**(code **)(local_a8[1] + 0x30))();
  }
  if ((-1 < (int)*local_120) && (*local_120 = *local_120 + -1, *local_120 == 0)) {
    (**(code **)(local_120[1] + 0x30))();
  }
  plVar11 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar15;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  return (code *)0x0;
}

